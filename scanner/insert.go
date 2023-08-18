package main

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
)

// Insert domains into Columbus.
// The goroutine is stopped by closing the LeafENtryChan in main().
func InsertWorker(doms <-chan string, wg *sync.WaitGroup) {

	defer wg.Done()

	for dom := range doms {

		if _, err := db.DomainsInsert(dom); err != nil {

			// Failed insert is fatal error. Dont want to miss any domain.
			fmt.Fprintf(os.Stderr, "Failed to write %s: %s\n", dom, err)

			// d is probably a TLD
			if errors.Is(err, fault.ErrGetPartsFailed) {
				continue
			}

			Cancel()
		}

		if err := db.RecordsUpdate(dom, true, false); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update records for %s: %s\n", dom, err)
		}
	}

}