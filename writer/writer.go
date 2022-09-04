package writer

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elmasy-com/columbus/reader"
	"github.com/elmasy-com/elnet/domain"
)

var (
	CheckUnique bool               // Check the uniqueness of the log entry
	OnlyFull    bool               // Only write full hostnames, skip sub/tld/domain
	WorkingDir  string             // Working directory
	BuffSize    int         = 2048 // Size of domain buffer
	NumFiles    int         = 1024 // Number of files to distribute entries
	domainChan  chan string        // Channel to send data from fetchDomain() to writer()
	fullFiles   []*os.File         // Files to write the full hostnames
	subFiles    []*os.File         // Files to write the subdomains
	domainFiles []*os.File         // Files to write the domains
	tldFile     *os.File           // File to write the TLDs
	running     bool               // Indicator that the writer is running
	m           sync.Mutex
)

func getRunning() bool {
	m.Lock()
	defer m.Unlock()
	return running
}

func setRunning(v bool) {
	m.Lock()
	defer m.Unlock()
	running = v
}

// writeFull writes the full domain name to the list.
func writeFull(wg *sync.WaitGroup, file *os.File, d []byte) {

	defer wg.Done()

	if CheckUnique {

		if exist, err := reader.IsExistDist(fullFiles, d); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check if %s is exist in full files: %s\n", d, err)
			return
		} else if exist {
			return
		}
	}

	// TODO: Check write error
	file.Write(d)
	file.Write([]byte{'\n'})
}

// writeSub writes the subdomain to the list.
func writeSub(wg *sync.WaitGroup, file *os.File, d []byte) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	v := domain.GetSub(d)

	if v == nil || bytes.Equal(v, []byte{'*'}) {
		return
	}

	if CheckUnique {

		if exist, err := reader.IsExistDist(subFiles, v); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check if %s is exist in sub files: %s\n", v, err)
			return
		} else if exist {
			return
		}
	}

	file.Write(v)
	file.Write([]byte{'\n'})
}

// wiretDomain writes the domain to the list.
func writeDomain(wg *sync.WaitGroup, file *os.File, d []byte) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	v := domain.GetDomain(d)
	if v == nil {
		return
	}

	if CheckUnique {

		if exist, err := reader.IsExistDist(domainFiles, v); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check if %s is exist in domain files: %s\n", v, err)
			return
		} else if exist {
			return
		}
	}

	file.Write(v)
	file.Write([]byte{'\n'})
}

// writeTLD writes TLD to the list.
func writeTLD(wg *sync.WaitGroup, file *os.File, d []byte) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	v := domain.GetTLD(d)
	if v == nil {
		return
	}

	if CheckUnique {

		if exist, err := reader.IsExistDist([]*os.File{tldFile}, v); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check if %s is exist in tld file: %s\n", v, err)
			return
		} else if exist {
			return
		}
	}

	file.Write(v)
	file.Write([]byte{'\n'})
}

// writer is used to write the lists, *this function is the bottleneck of this program* when used with CheckUnique.
// First it check the existence of the domain with isExist, then write it (if unique).
// This function stopped when the domainChan is closed in Close().
// TODO: How to make this faster!?
func writer() {

	setRunning(true)

	var (
		wg sync.WaitGroup
		db []byte
		i  int = 0
	)

	for d := range domainChan {

		if i >= NumFiles {
			//fmt.Printf("Waiting for WaitGroup in writer()\n")
			//wg.Wait()
			i = 0
		}

		db = []byte(d)

		wg.Add(4)

		go writeFull(&wg, fullFiles[i], db)
		go writeSub(&wg, subFiles[i], db)
		go writeDomain(&wg, domainFiles[i], db)
		go writeTLD(&wg, tldFile, db)

		wg.Wait()
		i++
	}

	wg.Wait()

	fmt.Printf("Writer closed!\n")
	setRunning(false)
}

// IsRunning checks whether the inner writer is running.
func IsRunning() bool {
	return getRunning()
}

// StartWriter starts the inner writer queue.
func StartWriter() error {

	// Dont start a new writer
	if getRunning() {
		return nil
	}

	if WorkingDir == "" {
		return fmt.Errorf("WorkingDir is empty")
	}

	// Create ./full with the files
	{
		path := fmt.Sprintf("%s/full", WorkingDir)

		if s, err := os.Stat(path); os.IsNotExist(err) {
			if errDir := os.Mkdir(path, 0755); errDir != nil {
				return fmt.Errorf("failed to create %s: %s", path, errDir)
			} else {
				fmt.Printf("Directory for full is created: %s\n", path)
			}
		} else if !s.IsDir() {
			return fmt.Errorf("failed to create %s: exist, but a directory", path)
		}

		for i := 0; i < NumFiles; i++ {

			path = fmt.Sprintf("%s/full/%d", WorkingDir, i)

			file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %s", path, err)
			}

			fullFiles = append(fullFiles, file)
		}
	}

	// Create ./sub and the files
	{
		path := fmt.Sprintf("%s/sub", WorkingDir)

		if s, err := os.Stat(path); os.IsNotExist(err) {
			if errDir := os.Mkdir(path, 0755); errDir != nil {
				return fmt.Errorf("failed to create %s: %s", path, errDir)
			} else {
				fmt.Printf("Directory for sub is created: %s\n", path)
			}
		} else if !s.IsDir() {
			return fmt.Errorf("failed to create %s: exist, but a directory", path)
		}

		for i := 0; i < NumFiles; i++ {

			path = fmt.Sprintf("%s/sub/%d", WorkingDir, i)
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %s", path, err)
			}

			subFiles = append(subFiles, file)
		}
	}

	// Create ./domain and the files
	{
		path := fmt.Sprintf("%s/domain", WorkingDir)

		if s, err := os.Stat(path); os.IsNotExist(err) {
			if errDir := os.Mkdir(path, 0755); errDir != nil {
				return fmt.Errorf("failed to create %s: %s", path, errDir)
			} else {
				fmt.Printf("Directory for sub is created: %s\n", path)
			}
		} else if !s.IsDir() {
			return fmt.Errorf("failed to create %s: exist, but a directory", path)
		}

		for i := 0; i < NumFiles; i++ {

			path := fmt.Sprintf("%s/domain/%d", WorkingDir, i)
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %s", path, err)
			}

			domainFiles = append(domainFiles, file)
		}
	}

	// Create tld file without folder, the numver of tlds are far less than the others
	{
		var err error
		path := fmt.Sprintf("%s/tld", WorkingDir)
		tldFile, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", path, err)
		}
		fmt.Printf("Path of the TLDs: %s\n", path)

	}

	domainChan = make(chan string, BuffSize)

	go writer()

	return nil
}

func Write(d string) {

	d = strings.ToLower(d)

	if domain.IsValid(d) {
		domainChan <- d
	}
}

// Close the inner writer.
// This function blocks until the oher function are not closed.
func Close() {

	if !getRunning() {
		return
	}

	close(domainChan)

	// Wait until domainChan is empty and write out every domain in buffer.
	for getRunning() {
		time.Sleep(1 * time.Second)
	}

	for i := range fullFiles {
		fullFiles[i].Close()
	}

	for i := range subFiles {
		subFiles[i].Close()
	}

	for i := range domainFiles {
		domainFiles[i].Close()
	}

	tldFile.Close()
}
