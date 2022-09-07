package reader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
)

// isExist opens file in path, iterate over the lines and returns whether the given entry is match with any of the lines through channel.
func isExist(ctx context.Context, wg *sync.WaitGroup, isFound chan<- bool, errChan chan<- error, path string, entry []byte) {

	defer wg.Done()

	var (
		found    bool
		err      error
		file     *os.File
		line     []byte
		lineLen  int
		entryLen = len(entry)
		buff     = make([]byte, 256)
		scanner  *bufio.Scanner
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
			lineLen = len(line)

			switch {
			case line[0] != entry[0]:
				continue
			case lineLen != entryLen:
				continue
			case lineLen > 1 && line[1] != entry[1]:
				continue
			case lineLen > 2 && line[2] != entry[2]:
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

// IsExistDist creates a goroutine for every file in files and
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

			// Every isExist() goroutine returned false.
			if scanned == l {
				goto exit
			}
		}
	}

exit:

	cancel()
	wg.Wait()
	close(isFound)
	close(errChan)

	return found, err
}
