package db

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatisticSchema struct {
	Date    int64         `bson:"date" json:"date"`
	Total   int64         `bson:"total" json:"total"`
	Updated int64         `bsn:"updated" json:"updated"`
	Valid   int64         `bson:"valid" json:"valid"`
	CTLogs  []CTLogSchema `bson:"ctlogs" json:"ctlogs"`
}

const (
	MaxStatisticsEntry = 500
)

// StatisticsCountTotal returns the total number of entries in "domain" collection.
func StatisticsCountTotal() (int64, error) {

	return Domains.CountDocuments(context.TODO(), bson.M{})
}

// StatisticsCountUpdated returns the total number of entries that updated in "domain" collection.
func StatisticsCountUpdated() (int64, error) {

	return Domains.CountDocuments(context.TODO(), bson.M{"updated": bson.M{"$exists": true}})
}

// StatisticsCountValid returns the total number of entries that has at least on valid record in the "records" field in "domain" collection.
func StatisticsCountValid() (int64, error) {

	return Domains.CountDocuments(context.TODO(), bson.M{"records": bson.M{"$exists": true}})
}

// StatisticsInsert get the stats and insert a new entry in the "statistics" collection.
//
// This function is **very** slow!
func StatisticsInsert() error {

	s := new(StatisticSchema)
	var err error

	s.Total, err = StatisticsCountTotal()
	if err != nil {
		return fmt.Errorf("failed to count total: %w", err)
	}

	s.Updated, err = StatisticsCountUpdated()
	if err != nil {
		return fmt.Errorf("failed to count updated: %w", err)
	}

	s.Valid, err = StatisticsCountValid()
	if err != nil {
		return fmt.Errorf("failed to count valid: %w", err)
	}

	s.CTLogs, err = CTLogsGets()
	if err != nil {
		return fmt.Errorf("failed to get CT logs: %w", err)
	}

	s.Date = time.Now().Unix()

	_, err = Statistics.InsertOne(context.TODO(), *s)

	return err
}

// StatisticsInsertWorker insert a new Statistic entry at the beginning and at a random time in an infinite loop.
//
// This function is designed to run as a goroutine in the background.
// The errors are printed to STDERR.
func StatisticsInsertWorker() {

	err := StatisticsInsert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert new statistic entry: %s\n", err)
	}

	for {

		time.Sleep(time.Duration(rand.Int63n(14400)) * time.Second)

		err := StatisticsInsert()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to insert new statistics entry: %s\n", err)
		}
	}
}

// StatisticsCleanWorker removes entries beyond MaxStatisticsEntry number.
//
// This function is designed to run as a goroutine in the background.
// The errors are printed to STDERR.
func StatisticsCleanWorker() {

	t := time.Tick(300 * time.Second)

	for range t {

		n, err := Statistics.CountDocuments(context.TODO(), bson.M{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "StatisticsRemoveOldEntries(): Failed to count total statistic entries: %s\n", err)
			continue
		}

		if n <= MaxStatisticsEntry {
			continue
		}

		i := 0

		cursor, err := Statistics.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"date": -1}))
		if err != nil {
			fmt.Fprintf(os.Stderr, "StatisticsRemoveOldEntries(): Failed to find statistic entries: %s\n", err)
			continue
		}

		for cursor.Next(context.TODO()) {

			if i <= MaxStatisticsEntry {
				i++
				continue
			}

			s := new(StatisticSchema)

			err = cursor.Decode(s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "StatisticsRemoveOldEntries(): Failed to decode: %s\n", err)
				continue
			}

			_, err := Statistics.DeleteOne(context.TODO(), *s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "StatisticsRemoveOldEntries(): Failed to remove entry (date: %d, total: %d, updated: %d, valid: %d): %s\n", s.Date, s.Total, s.Updated, s.Valid, err)
			}
		}

		err = cursor.Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "StatisticsRemoveOldEntries(): Cursor failed: %s\n", err)
		}

		cursor.Close(context.TODO())

	}
}

// StatisticsGetNewest returns the newest entry from the "statistics" collection.
func StatisticsGetNewest() (StatisticSchema, error) {

	s := new(StatisticSchema)

	err := Statistics.FindOne(context.TODO(), bson.M{}, options.FindOne().SetSort(bson.M{"date": -1})).Decode(s)

	return *s, err
}

// StatisticsGets returns every entry in the "statistics".
func StatisticsGets() ([]StatisticSchema, error) {

	cursor, err := Statistics.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		return nil, fmt.Errorf("failed to find: %w", err)
	}
	defer cursor.Close(context.TODO())

	r := make([]StatisticSchema, 0, MaxStatisticsEntry)

	for cursor.Next(context.TODO()) {

		s := new(StatisticSchema)

		err = cursor.Decode(s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %w", err)
		}

		r = append(r, *s)
	}

	err = cursor.Err()
	if err != nil {
		return nil, fmt.Errorf("cursor failed: %w", err)
	}

	return r, nil
}
