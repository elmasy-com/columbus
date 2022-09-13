package fetcher

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type uniqueEntry struct {
	domain string
	exist  bool
}

// isEntryExist check whether the given entry is exist in intries.
// Returns the matched entry index from entries or -1 if not found.
func isEntryExist(entries []uniqueEntry, entry string) int {

	for i := range entries {
		if entries[i].domain == entry {
			return i
		}
	}

	return -1
}

// genTempPath generate a hidden temp path for list file.
// Format: .list.<random>
func genTempPath(path string) string {

	rand.Seed(time.Now().UnixMicro())

	parts := strings.Split(path, "/")

	parts[len(parts)-1] = fmt.Sprintf(".%s.%d", parts[len(parts)-1], rand.Int63n(65536))
	return strings.Join(parts, "/")
}

// isExist opens file in path, iterate over the lines and set uniqueEntry.exist to true if the entry exist in file.
func isExist(ctx context.Context, path string, entries []uniqueEntry) error {

	var (
		buff = make([]byte, 256)
		n    int
	)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer(buff, len(buff)-1)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			return nil
		default:

			if n = isEntryExist(entries, scanner.Text()); n != -1 {
				entries[n].exist = true
			}
		}
	}

	err = scanner.Err()
	return err
}

// unique remove every duplicated entry from the given list.
// It creates a temporary file and renames it at the end.
func unique(ctx context.Context, path string) error {

	// TODO: Make unique context aware
	var (
		buff    = make([]byte, 256)
		entries = make([]uniqueEntry, 0, 10000)
	)

	tempPath := genTempPath(path)

	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	listFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer listFile.Close()

	scanner := bufio.NewScanner(listFile)

	scanner.Buffer(buff, len(buff)-1)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			// Terminate the loop ad remove the temp file
			tempFile.Close()
			err = os.Remove(tempPath)
			return err
		default:

			if isEntryExist(entries, scanner.Text()) != -1 {
				continue
			}

			entries = append(entries, uniqueEntry{domain: scanner.Text(), exist: false})

			// Continue until not read 10000 entries
			if len(entries) < 10000 {
				continue
			}

			err = isExist(ctx, tempPath, entries)
			if err != nil {
				return fmt.Errorf("failed to check entries: %s", err)
			}

			for i := range entries {
				if !entries[i].exist {
					tempFile.WriteString(entries[i].domain)
					tempFile.WriteString("\n")
				}
			}

			entries = make([]uniqueEntry, 0, 10000)
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %s", err)
	}

	// len(entries) < 10000, so the for loop exited before it checked/writed
	err = isExist(ctx, tempPath, entries)
	if err != nil {
		return fmt.Errorf("failed to check entries: %s", err)
	}

	for i := range entries {
		if !entries[i].exist {
			tempFile.WriteString(entries[i].domain)
			tempFile.WriteString("\n")
		}
	}

	// Defered close handle if file == nil
	tempFile.Close()
	listFile.Close()

	select {
	case <-ctx.Done():
		// Prevent to rename replace the list file
		tempFile.Close()
		err = os.Remove(tempPath)
		return err
	default:
		return os.Rename(tempPath, path)
	}
}
