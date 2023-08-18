package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/elnet/ctlog"
	"go.mongodb.org/mongo-driver/mongo"
)

type Status struct {
	Index int64
	Size  int64
	m     *sync.Mutex
}

func (s *Status) GetIndex() int64 {

	s.m.Lock()
	defer s.m.Unlock()

	return s.Index
}

func (s *Status) SetIndex(n int64) {

	s.m.Lock()
	defer s.m.Unlock()

	s.Index = n
}

func (s *Status) AddIndex(n int64) {

	s.m.Lock()
	defer s.m.Unlock()

	s.Index += int64(n)
}

func (s *Status) GetSize() int64 {

	s.m.Lock()
	defer s.m.Unlock()

	return s.Size
}

func (s *Status) SetSize(n int64) {

	s.m.Lock()
	defer s.m.Unlock()

	s.Size = n
}

func (s *Status) HasNew() bool {

	s.m.Lock()
	defer s.m.Unlock()

	return s.Size-s.Index > 0
}

var LogStat *Status

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

				LogStat.SetSize(s)
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

			err := db.CTLogsUpdate(Conf.LogName, LogStat.GetIndex(), LogStat.GetSize())
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

	LogStat.SetIndex(s.Index)

	LogStat.SetSize(s.Size)

	return nil
}
