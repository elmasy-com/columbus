package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Record is the schema used to store a record in Domain
type Record struct {
	Type  uint16 `bson:"type" json:"type"`
	Value string `bson:"value" json:"value"`
	Time  int64  `bson:"time" json:"time"`
}

var (
	RecordsUpdaterDomainChan         chan string
	internalRecordsUpdaterDomainChan chan string
	totalUpdated                     atomic.Uint64
	startTime                        time.Time
)

// increaseTotalUpdated add +1 to totalUpdated and print a status message.
func increaseTotalUpdated() {

	totalUpdated.Add(1)

	if totalUpdated.Load()%100000 == 0 {
		if totalUpdated.Load() != 0 {
			fmt.Printf("Updated %d domain records in %s\n", totalUpdated.Load(), time.Since(startTime))
		}
	}

}

// RecordsInsert insert (if not exist) or updates the "date" field for record with "type" t and "value" v.
// This function updates the "updated" field to the current time with DomainsUpdateUpdatedTime().
// If the same record found, updates the "time" field in element.
// If new record found, append it to the "records" field.
//
// Returns whether record with type t and value is a new record.
//
// If domain d is invalid, returns fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
func RecordsInsert(d string, t uint16, v string) (bool, error) {

	if !validator.Domain(d) {
		return false, fault.ErrInvalidDomain
	}

	p := dns.GetParts(dns.Clean(d))
	if p == nil || p.Domain == "" || p.TLD == "" {
		return false, fault.ErrGetPartsFailed
	}

	// "records" field should contain only one element with "type" t and "value" v.
	// Try to update first!
	// If MatchedCount is 0, the record with "type" t and "value" r[i] is new and the new record will be appended to the array.
	// If MatchedCount is 1, only one record is exist with "type" t and "value" v and the time for the element is updated.
	// If MatchedCount is > 1, duplicate record found, ERROR!
	filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}, {Key: "records.type", Value: t}, {Key: "records.value", Value: v}}

	up := bson.D{{Key: "$set", Value: bson.D{{Key: "records.$.time", Value: time.Now().Unix()}}}}

	result, err := Domains.UpdateOne(context.TODO(), filter, up)
	if err != nil {
		return false, err
	}

	if result.MatchedCount == 1 {
		return false, DomainsUpdateUpdatedTime(d)
	}

	// Append new record to "records"
	filter = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

	up = bson.D{{Key: "$addToSet", Value: bson.D{{Key: "records", Value: Record{Type: t, Value: v, Time: time.Now().Unix()}}}}}

	result, err = Domains.UpdateOne(context.TODO(), filter, up)
	if err != nil {
		return false, err
	}

	return result.ModifiedCount == 1, DomainsUpdateUpdatedTime(d)
}

// RecordsUpdate updates the records field for domain d if d is not update recently (in the previous hour).
// This function updates the "updated" field to the current time and the records in the database.
// If the same record found, updates the "time" field in element.
// If new record found, append it to the "records" field.
//
// Checks if d is a wildcard record before update.
//
// This function ignores the common DNS errors.
// If ignoreUpdated is true, ignore when was the last update based on the "updated" timestamp.
//
// If domain d is invalid, returns fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
func RecordsUpdate(d string, ignoreUpdated bool) error {

	d = dns.Clean(d)

	if !ignoreUpdated {

		updated, err := DomainsUpdatedRecently(d)
		if err != nil {
			return fmt.Errorf("failed to check if %s is updated recently: %w", d, err)
		}

		if updated {
			return nil
		}
	}

	records, err := dns.QueryAll(d)
	if err != nil && !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
		!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

		// Ignore common errors

		return fmt.Errorf("failed to update: %w", err)
	}

	if len(records) == 0 {
		return DomainsUpdateUpdatedTime(d)
	}

	for i := range records {
		_, err := RecordsInsert(d, records[i].Type, records[i].Value)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return nil
}

// recordsUpdaterRoutine reads from DomainChan and internalDomainChan
// and updates the FQDN coming from the channel.
func recordsUpdaterRoutine(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		var d string

		select {
		case dom := <-RecordsUpdaterDomainChan:
			d = dom

		case dom := <-internalRecordsUpdaterDomainChan:
			d = dom
		}

		if dns.HasSub(d) {

			increaseTotalUpdated()

			// d is a FQDN
			err := RecordsUpdate(d, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update DNS records for %s: %s\n", d, err)
			}

		} else {

			// If domain sent instead of FQDN, get every subdomain and updates it
			ds, err := DomainsLookupFull(d, -1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update DNS records for %s: %s\n", d, err)
				continue
			}

			for i := range ds {

				increaseTotalUpdated()

				err := RecordsUpdate(ds[i], false)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update DNS records for %s: %s\n", ds[i], err)
				}
			}
		}
	}
}

// RandomDomainUpdater is a function created to run as goroutine in the background.
// Select random entries (FQDNs) and send it to internalRecordsUpdaterDomainChan to update the records.
func RandomDomainUpdater(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		cursor, err := Domains.Aggregate(context.TODO(), bson.A{bson.M{"$sample": bson.M{"size": 1000}}})
		if err != nil {
			fmt.Fprintf(os.Stderr, "RandomDomainUpdater() failed to find toplist: %s\n", err)
			// Wait before the next try
			time.Sleep(600 * time.Second)
			continue
		}

		for cursor.Next(context.TODO()) {

			d := new(Domain)

			err = cursor.Decode(d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "RandomDomainUpdater() failed to find: %s\n", err)
				break
			}

			// TODO: Remove
			if d.Updated != 0 {
				continue
			}

			internalRecordsUpdaterDomainChan <- d.String()

		}

		if err = cursor.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "RandomDomainUpdater() cursor failed: %s\n", err)
		}

		cursor.Close(context.TODO())
	}
}

// TopListUpdater is a function created to run as goroutine in the background.
// Updates the domains and it subdomains in topList collection by sending every entries into internalRecordsUpdaterDomainChan.
// This function uses concurrent goroutines and print only/ignores any error.
func TopListUpdater(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		time.Sleep(time.Duration(rand.Intn(49) * int(time.Hour)))

		start := time.Now()

		cursor, err := TopList.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"count": -1}))
		if err != nil {
			fmt.Fprintf(os.Stderr, "TopListUpdater() failed to find toplist: %s\n", err)
			continue
		}

		for cursor.Next(context.TODO()) {

			d := new(TopListSchema)

			err = cursor.Decode(d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "TopListUpdater() failed to find: %s\n", err)
				break
			}

			ds, err := DomainsLookupFull(d.Domain, -1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "TopListUpdater() failed to lookup full for %s: %s\n", d.Domain, err)
				continue
			}

			for i := range ds {
				internalRecordsUpdaterDomainChan <- ds[i]
			}

		}

		if err = cursor.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "TopListUpdater() cursor failed: %s\n", err)
		}

		cursor.Close(context.TODO())
		fmt.Printf("TopListUpdater(): Finished updating topList in %s\n", time.Since(start))
	}
}

func RecordsUpdater(nworker int, chanSize int) {

	RecordsUpdaterDomainChan = make(chan string, chanSize)
	internalRecordsUpdaterDomainChan = make(chan string, chanSize)
	startTime = time.Now()

	wg := new(sync.WaitGroup)

	for i := 0; i < nworker; i++ {
		wg.Add(1)
		go recordsUpdaterRoutine(wg)
	}

	wg.Add(1)
	go RandomDomainUpdater(wg)

	wg.Add(1)
	go TopListUpdater(wg)

	wg.Wait()
}
