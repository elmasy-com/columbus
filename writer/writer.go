/*
 * writer is designed to Start() at the begining and Close() at the end.
 */
package writer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elmasy-com/elnet/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrRunning = errors.New("writer already running")

	domainChan chan string
	domainBuff []interface{}
	buffSize   = 1024

	client     *mongo.Client
	collection *mongo.Collection

	isRunnig bool
	m        sync.Mutex
)

func getRunning() bool {

	m.Lock()
	defer m.Unlock()

	return isRunnig
}

func setIsRunning(v bool) {

	m.Lock()
	defer m.Unlock()

	isRunnig = v
}

func writer() {

	setIsRunning(true)
	fmt.Printf("writer -> Started!\n")

	defer fmt.Printf("writer -> Stopped!\n")
	defer setIsRunning(false)

	i := 0
	dom := ""

	for d := range domainChan {

		dom = domain.GetDomain(d)
		if dom == "" {
			continue
		}

		domainBuff = append(domainBuff, bson.D{{"fqdn", d}, {"domain", dom}})

		i++

		if i >= buffSize {

			_, err := collection.InsertMany(context.TODO(), domainBuff)
			if err != nil {
				fmt.Printf("mongo -> Failed to insert many: %s\n", err)
				os.Exit(1)
			}

			// Reset domain buffer
			domainBuff = make([]interface{}, 0, buffSize)
			i = 0
		}
	}

	_, err := collection.InsertMany(context.TODO(), domainBuff)
	if err != nil {
		fmt.Printf("mongo -> Failed to insert many: %s\n", err)
	}
}

// Check if s is valid and send it to domainChan.
func Write(d string) {

	if domain.IsValid(d) {
		// domains are not case sensitive, store lower case domains to easily find duplicates later
		domainChan <- strings.ToLower(d)
	}

}

// Start connects to MongoDB and initialize the required chanel/buffer.
func Start() error {

	// Do not start again
	if getRunning() {
		return ErrRunning
	}

	var (
		err error
	)

	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return fmt.Errorf("failed to connect mongodb: %s", err)
	}

	collection = client.Database("columbus").Collection("domains")

	domainChan = make(chan string, buffSize)
	domainBuff = make([]interface{}, 0, buffSize)

	go writer()

	return nil
}

func Close() {

	if !getRunning() {
		return
	}

	close(domainChan)

	for getRunning() {
		time.Sleep(1 * time.Second)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Disconnect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "writer -> Failed to disconnect MongoDB: %s\n", err)
	} else {
		fmt.Printf("writer -> MongoDB disconnected!\n")
	}
}
