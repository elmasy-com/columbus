package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"go.mongodb.org/mongo-driver/bson"
)

// Record is the schema used to store a record in Domain
type Record struct {
	Type  uint16 `bson:"type" json:"type"`
	Value string `bson:"value" json:"value"`
	Time  int64  `bson:"time" json:"time"`
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

		// Skip empty records
		if records[i].Value == "" {
			continue
		}

		_, err := RecordsInsert(d, records[i].Type, records[i].Value)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return nil
}
