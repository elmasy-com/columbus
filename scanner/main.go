package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/elnet/ctlog"
	"github.com/g0rbe/slitu"
)

var (
	BuildDate   string
	BuildCommit string
	configPath  = flag.String("config", "", "Path to the config file")
	version     = flag.Bool("version", false, "Print current version")
	Cancel      context.CancelFunc
)

func main() {

	flag.Parse()

	if *version {
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", BuildCommit)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "-config is missing!\n")
		os.Exit(1)
	}

	fmt.Printf("Reading config file %s...\n", *configPath)

	err := ParseConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connecting to MongoDB...\n")
	err = db.Connect(Conf.MongoURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to MongoDB: %s\n", err)
		os.Exit(1)
	}
	defer db.Disconnect()

	LogIndex = new(atomic.Int64)
	LogSize = new(atomic.Int64)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	Cancel = cancel
	wg := new(sync.WaitGroup)
	domainChan := make(chan string)

	fmt.Printf("Loading previous LogStat...\n")
	err = LoadLogStat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load LogStat: %s\n", err)
		os.Exit(1)
	}
	if !LogHasNew() {
		fmt.Printf("%s progress: %d/%d (%.2f%%)\n", Conf.LogName, LogIndex.Load(), LogSize.Load(), float64(LogIndex.Load())/float64(LogSize.Load())*100)
	}

	wg.Add(1)
	go LogStatSizeUpdater(ctx, wg)

	wg.Add(1)
	go LogStatSaver(ctx, wg)

	for i := 0; i < Conf.InsertWorkers; i++ {
		wg.Add(1)
		go InsertWorker(domainChan, wg)
	}

infiniteLoop:
	for {

		select {
		case <-ctx.Done():
			break infiniteLoop
		default:

			if !LogHasNew() {
				// Nothing new, sleep a bit and retry
				slitu.Sleep(ctx, 5*time.Second)
				continue infiniteLoop
			} else {
				fmt.Printf("%s progress: %d/%d (%.2f%%)\n", Conf.LogName, LogIndex.Load(), LogSize.Load(), float64(LogIndex.Load())/float64(LogSize.Load())*100)
			}

			doms, n, err := ctlog.GetDomains(Conf.Log.URI, LogIndex.Load())
			if err != nil {

				switch {
				case strings.Contains(err.Error(), "NonFatalErrors"):
					// NonFatalErrors means failed to convert one entry, skip it and continue
					fmt.Fprintf(os.Stderr, "Non fatal error occurred while getting domains at index %d (continue from index %d): %s\n", LogIndex.Load()+n, LogIndex.Load()+n+1, err)
					// Add +1 to n to skip the failed entry
					n += 1
				case strings.Contains(err.Error(), "429 Too Many Requests"):
					fmt.Printf("Sleeping for 60 seconds because of too many request...\n")
					time.Sleep(60 * time.Second)
				default:
					fmt.Fprintf(os.Stderr, "Failed to get domains at index %d: %s\n", LogIndex.Load()+n, err)
					Cancel()
					break infiniteLoop
				}
			}

			for i := range doms {
				domainChan <- doms[i]
			}

			LogIndex.Add(n)
		}
	}

	fmt.Printf("Waiting to close...\n")
	close(domainChan)
	wg.Wait()
	fmt.Printf("Closed!\n")
	db.Disconnect()
	os.Exit(1)
}
