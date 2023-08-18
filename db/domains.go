package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/elmasy-com/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// DomainsLookup validate, Clean() and query the DB and returns a list subdomains only (eg.: "wwww", "mail").
// days specify, that the returned subdomain must had a valid record in the previous n days.
// If days is -1, every subdomain returned, including domains that does not have a record.
// If days is 0, return every subdomain that has a record regardless of the time.
//
// If d has a subdomain, removes it before the query.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns fault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsLookup(d string, days int) ([]string, error) {

	if !dns.IsValid(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return nil, fault.ErrGetPartsFailed
	}

	var filter primitive.D

	if days == 0 {
		// "records" field is exists
		filter = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "records", Value: bson.D{{Key: "$exists", Value: true}}}}
	} else if days == -1 {
		// Return every domain, the "records" filed doesnt matter
		filter = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}}
	} else if days > 0 {
		// Return every domain that has a record found in the last days days
		filter = bson.D{{Key: "domain", Value: p.Domain}, {Key: "tld", Value: p.TLD}, {Key: "records.time", Value: bson.D{{Key: "$gt", Value: time.Now().AddDate(0, 0, -1*days).Unix()}}}}
	} else {
		return nil, fault.ErrInvalidDays
	}

	// Use Find() to find every shard of the domain
	cursor, err := Domains.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find: %s", err)
	}
	defer cursor.Close(context.TODO())

	var subs []string

	for cursor.Next(context.TODO()) {

		r := new(FastDomain)

		err = cursor.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		subs = append(subs, r.Sub)
	}

	if err := cursor.Err(); err != nil {
		return subs, fmt.Errorf("cursor failed: %w", err)
	}

	return subs, nil
}

// DomainsLookupFull validate, Clean() and query the DB and returns a list full domains (eg.: "www.example.com", "mail.example.com").
// days specify, that the returned subdomain must had a valid record in the previous n days.
// If days is -1, every subdomain returned, including domains that does not have a record.
// If days is 0, return every subdomain that has a record regardless of the time.
//
// If d has a subdomain, removes it before the query.
//
// If d is invalid return fault.ErrInvalidDomain.
// If failed to get parts of d (eg.: d is a TLD), returns ault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsLookupFull(d string, days int) ([]string, error) {

	if !dns.IsValid(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return nil, fault.ErrGetPartsFailed
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

	var doms []string

	for cursor.Next(context.TODO()) {

		r := new(FastDomain)

		err = cursor.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %s", err)
		}

		doms = append(doms, r.String())
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
// NOTE: This function not validate adn Clean() d!
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
// If failed to get parts of d (eg.: d is a TLD), returns ault.ErrGetPartsFailed.
// If days if < -1, returns fault.ErrInvalidDays.
func DomainsRecords(d string, days int) ([]Record, error) {

	if !dns.IsValid(d) {
		return nil, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	p := dns.GetParts(d)
	if p == nil || p.Domain == "" || p.TLD == "" {
		return nil, fault.ErrGetPartsFailed
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
