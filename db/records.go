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

// RecordsUpdateUpdatedTime updated the "updated" timestamp to the current time.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
func RecordsUpdateUpdatedTime(d string) error {

	if !validator.Domain(d) {
		return fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return fault.ErrGetPartsFailed
	}

	filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

	up := bson.D{{Key: "$set", Value: bson.D{{Key: "updated", Value: time.Now().Unix()}}}}

	_, err := Domains.UpdateOne(context.TODO(), filter, up)

	return err
}

// RecordsUpdatedRecently check whether domain d is updated recently (in the previous hour).
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
func RecordsUpdatedRecently(d string) (bool, error) {

	if !validator.Domain(d) {
		return false, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return false, fault.ErrGetPartsFailed
	}

	filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

	dom := new(Domain)

	err := Domains.FindOne(context.TODO(), filter).Decode(dom)

	return dom.Updated > time.Now().Unix()-3600, err
}

// Update type t records for d.
// Check if domain d is a wildcard t type record.
// This function updates the DB.
func recordsUpdateRecord(d string, t uint16) error {

	if !validator.Domain(d) {
		return fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return fault.ErrGetPartsFailed
	}

	var (
		r   []string
		err error
	)

	// CHeck if domain has a t type wildcard record.
	wc, err := dns.IsWildcard(d, t)
	if err != nil {
		return err
	}
	if wc {
		return nil
	}

	switch t {

	case dns.TypeA:
		// A
		ips, err := dns.TryQueryA(d)
		if err != nil {
			return err
		}

		for i := range ips {
			r = append(r, ips[i].String())
		}

	case dns.TypeAAAA:
		// AAAA
		ips, err := dns.TryQueryAAAA(d)
		if err != nil {
			return err
		}

		for i := range ips {
			r = append(r, ips[i].String())
		}

	case dns.TypeCAA:
		// CAA
		caas, err := dns.TryQueryCAA(d)
		if err != nil {
			return err
		}

		for i := range caas {
			r = append(r, caas[i].String())
		}

	case dns.TypeCNAME:
		// CNAME
		r, err = dns.TryQueryCNAME(d)

	case dns.TypeDNAME:
		//DNAME
		v, err := dns.TryQueryDNAME(d)
		if err != nil {
			return err
		}

		r = append(r, v)

	case dns.TypeMX:
		// MX
		mxs, err := dns.TryQueryMX(d)
		if err != nil {
			return err
		}

		for i := range mxs {
			r = append(r, mxs[i].String())
		}

	case dns.TypeNS:
		// NS
		r, err = dns.TryQueryNS(d)

	case dns.TypeSOA:
		// SOA
		v, err := dns.TryQuerySOA(d)
		if err != nil {
			return err
		}

		if v != nil {
			r = append(r, v.String())
		}

	case dns.TypeSRV:
		// SRV
		srvs, err := dns.TryQuerySRV(d)
		if err != nil {
			return err
		}

		for i := range srvs {
			r = append(r, srvs[i].String())
		}
	case dns.TypeTXT:
		// TXT
		r, err = dns.TryQueryTXT(d)

	default:
		return fmt.Errorf("invalid type: %d", t)
	}

	if err != nil {
		return err
	}

	for i := range r {

		// "records" field should contain only one element with "type" t and "value" v.
		// Try to update first!
		// If MatchedCount is 0, the record with "type" t and "value" r[i] is new and the new record will be appended to the array.
		// If MatchedCount is 1, only one record is exist with "type" t and "value" v and the time for the element is updated.
		// If MatchedCount is > 1, duplicate record found, ERROR!
		filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}, {Key: "records.type", Value: t}, {Key: "records.value", Value: r[i]}}

		up := bson.D{{Key: "$set", Value: bson.D{{Key: "records.$.time", Value: time.Now().Unix()}}}}

		result, err := Domains.UpdateOne(context.TODO(), filter, up)
		if err != nil {
			return err
		}
		if result.MatchedCount > 1 {
			return fmt.Errorf("duplicate record found: %s", r[i])
		}
		if result.MatchedCount == 1 {
			continue
		}

		// Append new record to "records"
		filter = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

		up = bson.D{{Key: "$addToSet", Value: bson.D{{Key: "records", Value: Record{Type: t, Value: r[i], Time: time.Now().Unix()}}}}}

		_, err = Domains.UpdateOne(context.TODO(), filter, up)
		if err != nil {
			return err
		}
	}

	return nil
}

// RecordsUpdate updates the records field for domain d if d is not update recently (in the previous hour).
// This function updates the "updated" field to the current time and the records in the database.
// If the same record found, updates the "time" field in element.
// If new record found, append it to the "records" field.
//
// Checks if d is a wildcard record before update.
//
// If ignoreError is true, common DNS errors are ignored.
// If ignoreUpdated is true, ignore when was the last update based on the "updated" timestamp.
//
// If domain d is invalid, returns fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
func RecordsUpdate(d string, ignoreError bool, ignoreUpdated bool) error {

	if !ignoreUpdated {

		updated, err := RecordsUpdatedRecently(d)
		if err != nil {
			return fmt.Errorf("failed to check if %s is updated recently: %w", d, err)
		}

		if updated {
			return nil
		}
	}

	err := RecordsUpdateUpdatedTime(d)
	if err != nil {
		return fmt.Errorf("failed to update %s updated time: %w", d, err)
	}

	err = recordsUpdateRecord(d, dns.TypeA)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update A: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update A: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeAAAA)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update AAAA: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update AAAA: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeCAA)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update CAA: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update CAA: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeCNAME)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update CNAME: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update CNAME: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeDNAME)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update DNAME: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update DNAME: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeMX)
	if err != nil {
		if !ignoreError {
			return fmt.Errorf("failed to update MX: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update MX: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeNS)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update NS: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update NS: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeSOA)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update SOA: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update SOA: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeSRV)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update SRV: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update SRV: %w", err)
		}
	}

	err = recordsUpdateRecord(d, dns.TypeTXT)
	if err != nil {

		if !ignoreError {
			return fmt.Errorf("failed to update TXT: %w", err)
		}

		if !errors.Is(err, dns.ErrName) && !errors.Is(err, dns.ErrServerFailure) &&
			!os.IsTimeout(err) && !errors.Is(err, dns.ErrRefused) {

			return fmt.Errorf("failed to update TXT: %w", err)
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
			err := RecordsUpdate(d, true, false)
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

				err := RecordsUpdate(ds[i], true, false)
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
