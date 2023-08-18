package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/server/config"
	"github.com/elmasy-com/columbus/server/server"
)

var (
	BuildDate   string
	BuildCommit string
)

func main() {

	path := flag.String("config", "", "Path to the config file.")
	version := flag.Bool("version", false, "Print version informations.")
	flag.Parse()

	if *version {
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", BuildCommit)
		os.Exit(0)
	}

	if *path == "" {
		fmt.Fprintf(os.Stderr, "Path to the config file is missing!\n")
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Parsing config file...\n")
	if err := config.Parse(*path); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connecting to MongoDB...\n")
	if err := db.Connect(config.MongoURI); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to MongoDB: %s\n", err)
		os.Exit(1)
	}
	defer db.Disconnect()

	fmt.Printf("Starting db.StatisticsInsertWorker...\n")
	go db.StatisticsInsertWorker()

	fmt.Printf("Starting db.StatisticsCleanWorker...\n")
	go db.StatisticsCleanWorker()

	fmt.Printf("Starting RecordUpdater...\n")
	go db.RecordsUpdater(config.DomainWorker, config.DomainBuffer)

	fmt.Printf("Starting HTTP server...\n")
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("HTTP server stopped!\n")
	}
}
