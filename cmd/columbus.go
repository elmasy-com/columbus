package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/elmasy-com/columbus"
)

func main() {

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	dir := flag.String("dir", "", "Path to directory to store the files")

	flag.Parse()

	if *dir == "" {
		fmt.Fprintf(os.Stderr, "dir flag is empty\n")
		fmt.Fprintf(os.Stderr, "Use -h or -help for usage.\n")
		os.Exit(1)
	}

	*dir = strings.TrimSuffix(*dir, "/")

	columbus.Fetch(*dir)
}
