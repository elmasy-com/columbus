package db

import (
	"context"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Schema used in *topList* collection.
type TopListSchema struct {
	Domain string `bson:"domain" json:"domain"`
	Count  int    `bson:"count" json:"count"`
}

// TopListInsert inserts the given domain d to the *topList* database or increase the counter if exists.
// Checks if d is valid, do a Clean() and removes the subdomain from d.
//
// Returns true if d is new and inserted into the database.
// If domain is invalid or failed to remove the subdomain, returns fault.ErrInvalidDomain.
func TopListInsert(d string) (bool, error) {

	if !validator.Domain(d) {
		return false, fault.ErrInvalidDomain
	}

	d = dns.Clean(d)

	v := dns.GetDomain(d)
	if v == "" {
		return false, fault.ErrInvalidDomain
	}

	doc := bson.M{"domain": v}

	// UpdateOne will insert the document with $setOnInsert + $inc + upsert or do nothing
	res, err := TopList.UpdateOne(context.TODO(), doc, bson.M{"$setOnInsert": doc, "$inc": bson.M{"count": 1}}, options.Update().SetUpsert(true))

	return res.UpsertedCount != 0, err
}
