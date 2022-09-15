package finder

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/elmasy-com/elnet/domain"
	"github.com/elmasy-com/slices"
)

// getDirs returns the base
func getDirs(root string) ([]string, error) {

	dirs := make([]string, 0)

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not directory
		if !info.IsDir() {
			return nil
		}

		// Walk() list the root path, skip it
		if path == root {
			return nil
		}

		dirs = append(dirs, info.Name())
		return nil
	})

	return dirs, err
}

func readList(ctx context.Context, domainChan chan<- string, errChan chan<- error, path string, d string) {

	list := make([]string, 0)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		// select {
		// case <-ctx.Done():
		// 	fmt.Printf("time to run %s: %s\n", path, time.Since(s)) // TODO

		// 	return
		// default:
		if domain.GetDomain(scanner.Text()) == d {
			list = append(list, scanner.Text())
		}
		//}
	}

	select {
	case <-ctx.Done():
		// The channels are closed after cancel, prevent to write to closed channel

		//return
	default:

		for i := range list {
			domainChan <- list[i]

		}
		// Sending nil indicates that the goroutine is closed
		errChan <- scanner.Err()
	}
}

// TODO: Fix path for getDirs() <------------------------------------------------------
func Find(d string, workdir string) ([]string, error) {

	dirs, err := getDirs(workdir)
	if err != nil {
		return nil, fmt.Errorf("failed to get subdirectories: %s", err)
	}

	result := make([]string, 0)

	domainChan := make(chan string, 128)
	defer close(domainChan)

	errChan := make(chan error, len(dirs))
	defer close(errChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := range dirs {

		path := fmt.Sprintf("%s/%s/list", workdir, dirs[i])

		go readList(ctx, domainChan, errChan, path, d)
	}

	var closed int

	for {
		select {
		case err = <-errChan:
			if err != nil {
				return nil, err
			}

			closed++

			if closed == len(dirs) {
				return result, nil
			}

		case v := <-domainChan:
			//if domain.GetDomain(v) == d {
			if !slices.Contain(result, v) {
				result = append(result, v)
			}
			//}
		}
	}
}
