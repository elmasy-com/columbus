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

	"github.com/elmasy-com/columbus/config"
	"github.com/elmasy-com/columbus/fetcher"
	"github.com/elmasy-com/columbus/webserver"
)

func main() {

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flag.StringVar(&config.WorkingDir, "dir", "", "Path to directory to store the files")
	flag.IntVar(&config.Step, "step", 1000, "Number of logs queried at once. Must be greater than 0")

	flag.Parse()

	if config.WorkingDir == "" {
		fmt.Fprintf(os.Stderr, "dir flag is empty!\n")
		fmt.Fprintf(os.Stderr, "Use -h or -help for usage.\n")
		os.Exit(1)
	}

	config.WorkingDir = strings.TrimSuffix(config.WorkingDir, "/")

	if config.Step <= 0 {
		fmt.Fprintf(os.Stderr, "step must be greater than 0\n")
		fmt.Fprintf(os.Stderr, "Use -h or -help for usage.\n")
		os.Exit(1)
	}

	fetcher.Start()
	fmt.Printf("fetcher started!\n")

	webserver.Start()
	fmt.Printf("webserver started!\n")

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer sigCancel()

	for {
		select {
		case <-sigCtx.Done():
			fetcher.Close()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
