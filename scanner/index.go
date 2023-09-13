package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/elnet/ctlog"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	LogIndex *atomic.Int64
	LogSize  *atomic.Int64
)

func LogHasNew() bool {

	return LogSize.Load()-LogIndex.Load() > 0
}

func LogStatSizeUpdater(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {
		case <-ctx.Done():
			return
		default:

			s, err := ctlog.Size(Conf.Log.URI)
			if err != nil {
				if strings.Contains(err.Error(), "429 Too Many Requests") {
					time.Sleep(10 * time.Second)
				} else {
					fmt.Fprintf(os.Stderr, "Failed to update size: %s\n", err)
					Cancel()
					return
				}
			} else {

				LogSize.Store(s)

				time.Sleep(5 * time.Second)
			}
		}
	}
}

// LogStatSaver saves the Index periodically in a goroutine.
func LogStatSaver(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			err := db.CTLogsUpdate(Conf.LogName, LogIndex.Load(), LogSize.Load())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update LogStat in the database: %s\n", err)
				Cancel()
				return
			}
		}
	}
}

// LoadLogStat loads the index/size from the DB or defaiults to 0/0 if not found.
func LoadLogStat() error {

	s, err := db.CTLogsGet(Conf.LogName)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("failed to get: %w", err)
	}

	// If no document found, default to 0/0
	if s == nil {
		s = new(db.CTLogSchema)
	}

	LogIndex.Store(s.Index)

	LogSize.Store(s.Size)

	return nil
}
