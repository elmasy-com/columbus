package reader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// isExist opens file in path, iterate over the lines and returns whether the given entry is match with any of the lines.
func isExist(ctx context.Context, wg *sync.WaitGroup, isFound chan<- bool, errChan chan<- error, path string, entry []byte) {

	defer wg.Done()

	var (
		found   bool
		err     error
		file    *os.File
		line    []byte
		buff    = make([]byte, 0, 256)
		scanner *bufio.Scanner
	)

	file, err = os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		goto exit
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)

	scanner.Buffer(buff, len(buff)-1)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			goto exit
		default:

			line = scanner.Bytes()

			switch {
			case line[0] != entry[0]:
				continue
			case len(line) != len(entry):
				continue
			case bytes.Equal(line, entry):
				found = true
				goto exit
			}
		}
	}

	err = scanner.Err()

exit:
	isFound <- found
	errChan <- err
}

func IsExistDist(files []*os.File, entry []byte) (bool, error) {

	if files == nil {
		return false, fmt.Errorf("files is nil")
	}

	var (
		found       bool
		err         error
		l           = len(files)
		scanned     = 0 // Number of files scanned
		isFound     = make(chan bool, l)
		errChan     = make(chan error, l)
		ctx, cancel = context.WithCancel(context.Background())
		wg          sync.WaitGroup
		s           = time.Now() // TODO: Remove, only for debug/benchmark
	)

	for i := range files {
		wg.Add(1)
		go isExist(ctx, &wg, isFound, errChan, files[i].Name(), entry)
	}

	for {
		select {
		case err = <-errChan:
			if err != nil {
				goto exit
			}
		case found = <-isFound:
			if found {
				goto exit
			}

			scanned += 1
			if scanned == l {
				goto exit
			}
		}
	}

exit:

	// TODO: Remove, onlt for debug. Not accurate, just basic
	if time.Since(s) > 5*time.Second {
		fmt.Printf("Time to run IsExistDist(): %s\n", time.Since(s))
	}
	cancel()
	wg.Wait()
	close(isFound)
	close(errChan)

	return found, err
}
