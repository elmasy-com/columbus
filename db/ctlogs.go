package db

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Schema used in "ctlogs" collection
type CTLogSchema struct {
	Name  string `bson:"name" json:"name"`
	Index int64  `bson:"index" json:"index"`
	Size  int64  `bson:"size" json:"size"`
}

// CTLogsUpdate updates the stat for the CT log with name name in the *ctlogs* collection.
// The name is converted to lowercase.
func CTLogsUpdate(name string, index int64, size int64) error {

	name = strings.ToLower(name)

	_, err := CTLogs.UpdateOne(context.TODO(), bson.D{{Key: "name", Value: name}}, bson.D{{Key: "$set", Value: bson.D{{Key: "index", Value: index}, {Key: "size", Value: size}}}}, options.Update().SetUpsert(true))

	return err
}

// CTLogsGet returns the stat for CT log with name name from the *ctlogs* collection.
// The name is converted to lowercase.
func CTLogsGet(name string) (*CTLogSchema, error) {

	name = strings.ToLower(name)

	s := new(CTLogSchema)

	err := CTLogs.FindOne(context.TODO(), bson.D{{Key: "name", Value: name}}).Decode(s)

	return s, err
}

// CTLogsGets returns every entry from the "ctlogs" database.
func CTLogsGets() ([]CTLogSchema, error) {

	cursor, err := CTLogs.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to find: %w", err)
	}
	defer cursor.Close(context.TODO())

	scs := make([]CTLogSchema, 0)

	for cursor.Next(context.TODO()) {

		sc := new(CTLogSchema)

		err = cursor.Decode(sc)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %w", err)
		}

		scs = append(scs, *sc)
	}

	return scs, cursor.Err()
}
