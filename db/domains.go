package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/elmasy-com/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FastDomain is a the schema used in Lookup() to ignore the Records field.
type FastDomain struct {
	Domain string `bson:"domain" json:"domain"`
	TLD    string `bson:"tld" json:"tld"`
	Sub    string `bson:"sub" json:"sub"`
}

// Returns the full hostname (eg.: sub.domain.tld).
func (d *FastDomain) String() string {

	if d.Sub == "" {
		return strings.Join([]string{d.Domain, d.TLD}, ".")
	} else {
		return strings.Join([]string{d.Sub, d.Domain, d.TLD}, ".")
	}
}

// Domains is the schema used in the "domains" collection.
type Domain struct {
	Domain  string   `bson:"domain" json:"domain"`
	TLD     string   `bson:"tld" json:"tld"`
	Sub     string   `bson:"sub" json:"sub"`
	Updated int64    `bson:"updated" json:"updated"`
	Records []Record `bson:"records,omitempty" json:"records,omitempty"`
}

// Returns the full hostname (eg.: sub.domain.tld).
func (d *Domain) String() string {

	if d.Sub == "" {
		return strings.Join([]string{d.Domain, d.TLD}, ".")
	} else {
		return strings.Join([]string{d.Sub, d.Domain, d.TLD}, ".")
	}
}

// Returns the domain and tld only (eg.: domain.tld)
func (d *Domain) FullDomain() string {
	return strings.Join([]string{d.Domain, d.TLD}, ".")
}

// DomainsInsert inserts the given domain d to the *domains* database.
// Checks if d is valid, do a Clean() and then splits into sub|domain|tld parts.
//
// Returns true if d is new and inserted into the database.
// If domain is invalid, returns fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns ault.ErrGetPartsFailed.
//
// NOTE: Use RecordsUpdate() after Insert()!
func DomainsInsert(d string) (bool, error) {

	if !validator.Domain(d) {
		return false, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return false, fault.ErrGetPartsFailed
	}

	doc := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

	// UpdateOne will insert the document with $setOnInsert + upsert or do nothing
	res, err := Domains.UpdateOne(context.TODO(), doc, bson.M{"$setOnInsert": doc}, options.Update().SetUpsert(true))
	if err != nil {
		return false, fmt.Errorf("failed to update: %w", err)
	}

	return res.UpsertedCount != 0, nil
}

// DomainsInsertWithRecord inserts the given domain d to the *domains* database IF d has at least one valid record.
// Checks if d is valid, do a Clean() and search for records. If found at least one valid record, insert into the database.
// This function always updates the "updated" field, regardless of the records.
//
// This function returns if domain d is updated recently.
// This function ignores common DNS errors (eg.: NXDOMAIN).
//
// If domain is invalid, returns fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns ault.ErrGetPartsFailed.
func DomainsInsertWithRecord(d string, ignoreUpdated bool) error {

	d = dns.Clean(d)

	if !ignoreUpdated {

		// DomainsUpdatedRecently check if d is valid.
		updated, err := DomainsUpdatedRecently(d)
		if err != nil {
			return err
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
		return nil
	}

	_, err = DomainsInsert(d)
	if err != nil {
		return fmt.Errorf("failed to insert domain: %s", err)
	}

	for i := range records {

		_, err = RecordsInsert(d, records[i].Type, records[i].Value)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return nil
}

// DomainsLookup validate, Clean() and query the DB and returns a list subdomains only (eg.: "wwww", "mail").
// days specify, that the returned subdomain must had a valid record in the previous n days.
// If days is -1, every subdomain returned, including domains that does not have a record.
// If days is 0, return every subdomain that has a record regardless of the time.
//
// If d has a subdomain, removes it before the query.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsLookup(d string, days int) ([]string, error) {

	ds, err := DomainsDomains(d, days)
	if err != nil {
		return nil, err
	}

	doms := make([]string, 0, len(ds))

	for i := range ds {
		doms = append(doms, ds[i].Sub)
	}

	return doms, nil
}

// DomainsLookupFull validate, Clean() and query the DB and returns a list full domains (eg.: "www.example.com", "mail.example.com").
// days specify, that the returned subdomain must had a valid record in the previous n days.
// If days is -1, every subdomain returned, including domains that does not have a record.
// If days is 0, return every subdomain that has a record regardless of the time.
//
// If d has a subdomain, removes it before the query.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsLookupFull(d string, days int) ([]string, error) {

	ds, err := DomainsDomains(d, days)
	if err != nil {
		return nil, err
	}

	doms := make([]string, 0, len(ds))

	for i := range ds {
		doms = append(doms, ds[i].String())
	}

	return doms, nil
}

// DomainsDomains validate, Clean() and query the DB and returns a list of Domains.
// days specify, that the returned Domain must have a valid record in the previous n days.
// If days is -1, every Domain returned, including Domains that does not have a record.
// If days is 0, return every Domain that has a record regardless of the time.
//
// If d has a subdomain, removes it before the query.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsDomains(d string, days int) ([]Domain, error) {

	if !dns.IsValid(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.TLD == "" {
		return nil, fault.ErrGetPartsFailed
	}
	if p.Domain == "" {
		return nil, fault.ErrTLDOnly
	}

	var doc primitive.D

	if days == 0 {
		// "records" field is exists
		doc = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "records", Value: bson.D{{Key: "$exists", Value: true}}}}
	} else if days == -1 {
		// Return every domain, the "records" filed doesnt matter
		doc = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}}
	} else if days > 0 {
		// Return every domain that has a record found in the last days days
		doc = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "records.time", Value: bson.D{{Key: "$gt", Value: time.Now().AddDate(0, 0, -1*days).Unix()}}}}
	} else {
		return nil, fault.ErrInvalidDays
	}

	// Use Find() to find every shard of the domain
	cursor, err := Domains.Find(context.TODO(), doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find: %s", err)
	}
	defer cursor.Close(context.TODO())

	var doms []Domain

	for cursor.Next(context.TODO()) {

		r := new(Domain)

		err = cursor.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		doms = append(doms, *r)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor failed: %w", err)
	}

	return doms, nil
}

// DomainsTLD query the DB and returns a list of TLDs for the given domain d (eg.: "com", "org").
//
// Domain d must be a valid Second Level Domain (eg.: "example").
//
// NOTE: This function not validate and Clean() d!
func DomainsTLD(d string) ([]string, error) {

	// Use Find() to find every shard of the domain
	cursor, err := Domains.Find(context.TODO(), bson.M{"domain": d})
	if err != nil {
		return nil, fmt.Errorf("failed to find: %s", err)
	}
	defer cursor.Close(context.TODO())

	var tlds []string

	for cursor.Next(context.TODO()) {

		var r FastDomain

		err = cursor.Decode(&r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		tlds = slices.AppendUnique(tlds, r.TLD)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor failed: %w", err)
	}

	return tlds, nil
}

// DomainsStarts query the DB and returns a list of Second Level Domains (eg.: "reddit", "redditmedia") that starts with d.
//
// Domain d must be a valid Second Level Domain (eg.: "example").
// This function validate with IsValidSLD() and Clean().
//
// Returns fault.ErrInvalidDomain is d is not a valid Second Level Domain.
func DomainsStarts(d string) ([]string, error) {

	if !dns.IsValidSLD(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	filter := bson.M{"domain": bson.M{"$regex": fmt.Sprintf("^%s", d)}}

	// Use Find() to find every shard of the domain
	cursor, err := Domains.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find: %s", err)
	}
	defer cursor.Close(context.TODO())

	var domains []string

	for cursor.Next(context.TODO()) {

		var r FastDomain

		err = cursor.Decode(&r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		domains = slices.AppendUnique(domains, r.Domain)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor failed: %w", err)
	}

	return domains, nil
}

// DomainsRecords query the DB and returns a list Record.
// days specify, that the returned record must be updated in the previous n days.
// If days is 0 or -1, return every record regardless of the time.
//
// Returns records for the exact domain d.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsRecords(d string, days int) ([]Record, error) {

	if !dns.IsValid(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.TLD == "" {
		return nil, fault.ErrGetPartsFailed
	}
	if p.Domain == "" {
		return nil, fault.ErrTLDOnly
	}

	var doc primitive.D

	if days == 0 || days == -1 {
		// "records" field is/must exists
		doc = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}, {Key: "records", Value: bson.D{{Key: "$exists", Value: true}}}}
	} else if days > 0 {
		// Return every domain that has a record found in the last days days
		doc = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}, {Key: "records.time", Value: bson.D{{Key: "$gt", Value: time.Now().AddDate(0, 0, -1*days).Unix()}}}}
	} else {
		return nil, fault.ErrInvalidDays
	}

	// Use Find() to find every shard of the domain
	cursor, err := Domains.Find(context.TODO(), doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find: %s", err)
	}
	defer cursor.Close(context.TODO())

	var records = make([]Record, 0)

	for cursor.Next(context.TODO()) {

		r := new(Domain)

		err = cursor.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		records = append(records, r.Records...)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor failed: %w", err)
	}

	return records, nil
}

// DomainsUpdateUpdatedTime updated the "updated" timestamp to the current time to domain d.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
func DomainsUpdateUpdatedTime(d string) error {

	if !validator.Domain(d) {
		return fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.TLD == "" {
		return fault.ErrGetPartsFailed
	}
	if p.Domain == "" {
		return fault.ErrTLDOnly
	}

	filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}}

	up := bson.D{{Key: "$set", Value: bson.D{{Key: "updated", Value: time.Now().Unix()}}}}

	_, err := Domains.UpdateOne(context.TODO(), filter, up)

	return err
}

// DomainsUpdatedRecently check whether domain d is updated recently (in the previous 12 hours).
//
// Return false, nil if d is not exists in the database (ignore mongo.ErrNoDocuments).
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d because of d is just a TLD, returns fault.ErrTLDOnly.
// If failed to get parts of d, returns fault.ErrGetPartsFailed.
func DomainsUpdatedRecently(d string) (bool, error) {

	if !validator.Domain(d) {
		return false, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.TLD == "" {
		return false, fault.ErrGetPartsFailed
	}
	if p.Domain == "" {
		return false, fault.ErrTLDOnly
	}

	filter := bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "sub", Value: p.Sub}, {Key: "updated", Value: bson.M{"$gt": time.Now().UTC().Add(-12 * time.Hour).Unix()}}}

	dom := new(Domain)

	err := Domains.FindOne(context.TODO(), filter).Decode(dom)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}

	return false, err
}
