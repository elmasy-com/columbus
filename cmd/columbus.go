package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elmasy-com/columbus/closer"
	"github.com/elmasy-com/columbus/fetcher"
	"github.com/elmasy-com/columbus/writer"
)

func main() {

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flag.StringVar(&writer.WorkingDir, "dir", "", "Path to directory to store the files")
	flag.BoolVar(&writer.CheckUnique, "check", false, "Check the uniqueness of the log entry")
	flag.BoolVar(&writer.OnlyFull, "full", false, "Only write full hostnames")

	flag.Parse()

	if writer.WorkingDir == "" {
		fmt.Fprintf(os.Stderr, "dir flag is empty!\n")
		fmt.Fprintf(os.Stderr, "Use -h or -help for usage.\n")
		os.Exit(1)
	}

	writer.WorkingDir = strings.TrimSuffix(writer.WorkingDir, "/")

	if err := writer.StartWriter(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start writer: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("writer started!\n")

	fetcher.Start()
	fmt.Printf("fetcher started!\n")

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer sigCancel()

	for {
		select {
		case <-sigCtx.Done():
			closer.Closer()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
