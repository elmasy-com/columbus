package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateType uint8

const (
	InsertNewDomain UpdateType = iota
	UpdateExistingDomain
)

// UpdateableDomain used to distinguish domain coming from /api/insert and domains coming from updater functions.
type UpdateableDomain struct {
	Domain string
	Type   UpdateType
}

var (
	UpdaterChan  chan UpdateableDomain
	updaterLimit int
)

// recordsUpdaterRoutine reads from DomainChan and internalDomainChan
// and updates the FQDN coming from the channel.
func updaterWorker(wg *sync.WaitGroup) {

	defer wg.Done()

	for dom := range UpdaterChan {

		var err error

		switch dom.Type {
		case InsertNewDomain:
			fmt.Printf("updater: inserting %s\n", dom.Domain)
			err = DomainsInsertWithRecord(dom.Domain, false)
		case UpdateExistingDomain:
			err = RecordsUpdate(dom.Domain, false)
		default:
			err = fmt.Errorf("invalid UpdateType: %d", dom.Type)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update DNS records for %s in updaterWorker(): %s\n", dom.Domain, err)
		}
	}
}

// oldDomainUpdater is a function created to run as goroutine in the background.
// Updates the old records, that not updated in the last 30 days
func oldDomainUpdater(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
				bson.M{"updated": bson.M{"$lt": time.Now().Add(-720 * time.Hour).Unix()}},
				bson.M{"updated": bson.M{"$exists": false}},
			}},
			}}}

		sampleStage := bson.D{{Key: "$sample", Value: bson.M{"size": 1000}}}

		// Get domains that not updated in the last 30 days
		cursor, err := Domains.Aggregate(context.TODO(), mongo.Pipeline{matchStage, sampleStage})

		//cursor, err := Domains.Find(context.TODO(), bson.M{"$or": bson.A{bson.M{"updated": bson.M{"$lt": time.Now().Add(-720 * time.Hour).Unix()}}, bson.M{"updated": bson.M{"$exists": false}}}})
		if err != nil {
			fmt.Fprintf(os.Stderr, "oldDomainUpdater() failed to find toplist: %s\n", err)
			// Wait before the next try
			time.Sleep(600 * time.Second)
			continue
		}

		for cursor.Next(context.TODO()) {

			d := new(Domain)

			err = cursor.Decode(d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "oldDomainUpdater() failed to find: %s\n", err)
				break
			}

			for len(UpdaterChan) > updaterLimit {
				time.Sleep(60 * time.Second)
			}

			UpdaterChan <- UpdateableDomain{Domain: d.String(), Type: UpdateExistingDomain}

		}

		if err = cursor.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "OldDomainUpdater() cursor failed: %s\n", err)
		}

		cursor.Close(context.TODO())
	}
}

// topListUpdater is a function created to run as goroutine in the background.
// Updates the domains and it subdomains in topList collection by sending every entries into internalRecordsUpdaterDomainChan.
// This function uses concurrent goroutines and print only/ignores any error.
func topListUpdater(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		sampleStage := bson.D{{Key: "$sample", Value: bson.M{"size": 1000}}}

		cursor, err := TopList.Aggregate(context.TODO(), mongo.Pipeline{sampleStage})
		//cursor, err := TopList.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"count": -1}))
		if err != nil {
			fmt.Fprintf(os.Stderr, "topListUpdater() failed to find toplist: %s\n", err)
			continue
		}

		for cursor.Next(context.TODO()) {

			d := new(TopListSchema)

			err = cursor.Decode(d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "topListUpdater() failed to find: %s\n", err)
				break
			}

			ds, err := DomainsLookupFull(d.Domain, -1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "topListUpdater() failed to lookup full for %s: %s\n", d.Domain, err)
				continue
			}

			for i := range ds {

				for len(UpdaterChan) > updaterLimit {
					time.Sleep(60 * time.Second)
				}

				UpdaterChan <- UpdateableDomain{Domain: ds[i], Type: UpdateExistingDomain}
			}

		}

		if err = cursor.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "topListUpdater() cursor failed: %s\n", err)
		}

		cursor.Close(context.TODO())
	}
}

func RecordsUpdater(nworker int, chanSize int) {

	UpdaterChan = make(chan UpdateableDomain, chanSize)

	updaterLimit = chanSize / 2

	wg := new(sync.WaitGroup)

	for i := 0; i < nworker; i++ {
		wg.Add(1)
		go updaterWorker(wg)
	}

	wg.Add(1)
	go oldDomainUpdater(wg)

	wg.Add(1)
	go topListUpdater(wg)

	wg.Wait()
}
