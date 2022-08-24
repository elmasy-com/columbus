package columbus

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/elmasy-com/elnet/domain"
	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
)

type log struct {
	URI   string // URI of the log
	Name  string // "nickname" of the log
	Index int    // Index to start with, updated from the .index file if present
}

var (
	DEBUG_MODE = false

	// From https://www.gstatic.com/ct/log_list/log_list.json
	URIS = []log{
		{"https://ct.googleapis.com/logs/argon2022/", "argon2022", 0},
		{"https://ct.googleapis.com/logs/argon2023/", "argon2023", 0},
		{"https://ct.googleapis.com/logs/xenon2022/", "xenon2022", 0},
		{"https://ct.googleapis.com/logs/xenon2023/", "xenon2023", 0},
		{"https://ct.googleapis.com/icarus/", "icarus", 0},
		{"https://ct.googleapis.com/pilot/", "pilot", 0},
		{"https://ct.googleapis.com/rocketeer/", "rocketeer", 0},
		{"https://ct.googleapis.com/skydiver/", "skydiver", 0},
		{"https://ct.cloudflare.com/logs/nimbus2022/", "nimbus2022", 0},
		{"https://ct.cloudflare.com/logs/nimbus2023/", "nimbus2023", 0},
		{"https://ct1.digicert-ct.com/log/", "digicert-ct1", 0},
		{"https://ct2.digicert-ct.com/log/", "digicert-ct2", 0},
		{"https://yeti2022.ct.digicert.com/log/", "yeti2022", 0},
		{"https://yeti2022-2.ct.digicert.com/log/", "yeti2022-2", 0},
		{"https://yeti2023.ct.digicert.com/log/", "yeti2023", 0},
		{"https://nessie2022.ct.digicert.com/log/", "nessie2022", 0},
		{"https://nessie2023.ct.digicert.com/log/", "nessie2023", 0},
		{"https://sabre.ct.comodo.com/", "sabre", 0},
		{"https://mammoth.ct.comodo.com/", "mammoth", 0},
		{"https://oak.ct.letsencrypt.org/2019/", "oak2019", 0},
		{"https://oak.ct.letsencrypt.org/2020/", "oak2020", 0},
		{"https://oak.ct.letsencrypt.org/2021/", "oak2021", 0},
		{"https://oak.ct.letsencrypt.org/2022/", "oak2022", 0},
		{"https://oak.ct.letsencrypt.org/2023/", "oak2023", 0},
		{"https://ct.trustasia.com/log2022/", "trustasia2022", 0},
		{"https://ct.trustasia.com/log2023/", "trustasia2023", 0},
	}

	fullFile       *os.File // File to write the full hostnames
	fullFilePath   string   // Path of the full hostnames
	subFile        *os.File // File to write the subdomains
	subFilePath    string   // Path of the subdomains
	domainFile     *os.File // File to write the domains
	domainFilePath string   // Path of the domains
	tldFile        *os.File // File to write the TLDs
	tldFilePath    string   // Path of the TLDs

	domainChan chan string // Channel to send data from fetchDomain() to writer()

	CheckUnique bool // Check the uniqueness of the log entry
	OnlyFull    bool // Only write full hostnames, skip sub/tld/domain
)

func updateURI(name string, index int) error {

	for i := range URIS {

		if name == URIS[i].Name {
			URIS[i].Index = index
			return nil
		}
	}

	return fmt.Errorf("URI for %s not found", name)
}

// Load the log.Index from the svaed file.
func readIndex(dir string) error {

	if _, err := os.Stat(dir + "/columbus.index"); os.IsNotExist(err) {
		return nil
	}

	out, err := os.ReadFile(dir + "/columbus.index")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")

	for i := range lines {

		fields := strings.Split(lines[i], "=")

		if len(fields) != 2 {
			continue
		}

		ind, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index %s to int for %s", fields[1], fields[0])
		}

		updateURI(fields[0], ind)
	}

	return nil
}

// Save url.Index.
func saveIndex(dir string) error {

	file, err := os.OpenFile(dir+"/columbus.index", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := range URIS {

		// Skip logs, thats not working/no progress.
		if URIS[i].Index == 0 {
			continue
		}

		_, err = file.WriteString(fmt.Sprintf("%s=%d\n", URIS[i].Name, URIS[i].Index))
		if err != nil {
			return err
		}
	}

	return nil
}

// backgroundSave saves the indexes of the URIS every 10 minute.
// In case of any unexpected termination.
// signal context will terminate this, but a last save will happen at the end of Fetch()
func backgroundSave(ctx context.Context, dir string) {

	waitUntil := time.Now().Add(1 * time.Minute)

	for {

		select {
		case <-ctx.Done():
			fmt.Printf("Closing background saver...\n")
			return
		default:

			if time.Now().After(waitUntil) {
				if err := saveIndex(dir); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to save indexes: %s\n", err)
				}

				waitUntil = time.Now().Add(1 * time.Minute)
			}

			time.Sleep(10 * time.Second)
		}
	}
}

// isExist opens file in path, iterate over the lines and returns whether the given entry is match with any of the lines.
func isExist(path string, entry []byte) (bool, error) {

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buff := make([]byte, 0, 256)

	scanner := bufio.NewScanner(file)

	scanner.Buffer(buff, len(buff)-1)

	for scanner.Scan() {
		if bytes.Compare(scanner.Bytes(), entry) == 0 {
			return true, nil
		}
	}

	return false, nil
}

// writeFull writes the full domain name to the list.
func writeFull(wg *sync.WaitGroup, d []byte) {

	defer wg.Done()

	if CheckUnique {

		if exist, err := isExist(fullFilePath, d); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check the existence of %s in writeFull(): %s\n", d, err)
			return
		} else if exist {
			return
		}
	}

	fullFile.Write(d)
	fullFile.Write([]byte{'\n'})
}

// writeSub writes the subdomain to the list.
func writeSub(wg *sync.WaitGroup, d string) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	s := domain.GetSub(d)

	if s == "" || s == "*" {
		return
	}

	if CheckUnique {

		if exist, err := isExist(subFilePath, []byte(s)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check the existence of %s in writeSub(): %s\n", s, err)
			return
		} else if exist {
			return
		}
	}

	subFile.WriteString(s)
	subFile.WriteString("\n")
}

// wiretDomain writes the domain to the list.
func writeDomain(wg *sync.WaitGroup, d string) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	v := domain.GetDomain(d)

	if v == "" {
		return
	}

	if CheckUnique {

		if exist, err := isExist(domainFilePath, []byte(v)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check the existence of %s in writeDomain(): %s\n", v, err)
			return
		} else if exist {
			return
		}
	}

	domainFile.WriteString(v)
	domainFile.WriteString("\n")
}

// writeTLD writes TLD to the list.
func writeTLD(wg *sync.WaitGroup, d string) {

	defer wg.Done()

	if OnlyFull {
		return
	}

	v := domain.GetTLD(d)

	if v == "" {
		return
	}

	if CheckUnique {
		if exist, err := isExist(tldFilePath, []byte(v)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check the existence of %s in writeTLD(): %s\n", v, err)
			return
		} else if exist {
			return
		}
	}

	tldFile.WriteString(v)
	tldFile.WriteString("\n")
}

// writer is used to write the lists, *this function is the bottleneck of this program* when used with CheckUnique.
// First it check the existence of the domain with isExist, then write it (if unique).
// This function stopped when the domainChan is closed at the end of the Fetch().
// TODO: How to make this faster!?
func writer(wg *sync.WaitGroup) {

	defer wg.Done()

	for d := range domainChan {

		var wg sync.WaitGroup

		wg.Add(4)

		go writeFull(&wg, []byte(d))
		go writeSub(&wg, d)
		go writeDomain(&wg, d)
		go writeTLD(&wg, d)

		wg.Wait()
	}

	fmt.Printf("Closing writer...\n")
}

func fetcher(wg *sync.WaitGroup, ctx context.Context, l *log) {
	//func fetcher(l log) {

	defer wg.Done()

	if l.Index > 0 {
		fmt.Printf("%s -> Continuing from index %d\n", l.Name, l.Index)
	}

	logClient, err := client.New(
		l.URI,
		&http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   30 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				MaxIdleConnsPerHost:   10,
				DisableKeepAlives:     false,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		jsonclient.Options{UserAgent: "ct-go-scanlog/1.0"},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s -> Failed to create new client: %s\n", l.Name, err)
		return
	}

	sth, err := logClient.GetSTH(context.TODO())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s -> Failed to get SignedTreeHead: %s\n", l.Name, err)
		return
	}

	fmt.Printf("%s -> Number of logs: %d\n", l.Name, sth.TreeSize)

	// Number of seconds to wait in case of "429 Too Many Requests"
	toWait := 0

	// Logs MAY return fewer than the number of leaves requested. Only complete
	// if we actually got all the leaves we were expecting.
	// See more: https://github.com/google/certificate-transparency-go/blob/52d94d8cbab94d6698621839ab1a439d17ebbfb2/scanner/fetcher.go#L263
	for uint64(l.Index) < sth.TreeSize {

		time.Sleep(time.Duration(toWait) * time.Second)

		select {
		case <-ctx.Done():
			fmt.Printf("Shuting down %s log fetcher at index %d...\n", l.Name, l.Index)
			return
		default:

			if DEBUG_MODE {
				fmt.Printf("%s -> New fetch start with index %d-%d\n", l.Name, l.Index, l.Index+100)
			}

			entries, err := logClient.GetRawEntries(context.TODO(), int64(l.Index), int64(l.Index+100))
			if err != nil {

				// Sleep 1 minutes if too many request
				// TODO: context have to wait 1 minute to terminate
				if strings.Contains(err.Error(), "429 Too Many Requests") {
					toWait += 10
					fmt.Fprintf(os.Stderr, "%s -> Failed to get raw entries: %s. Waiting for %d seconds to retry...\n", l.Name, err, toWait)
					continue
				}

				fmt.Fprintf(os.Stderr, "%s -> Failed to get raw entries: %s\n", l.Name, err)
				return
			}

			if entries == nil {
				fmt.Fprintf(os.Stderr, "%s -> entries is nil", l.Name)
				return
			}

			if DEBUG_MODE {
				fmt.Printf("%s -> Got %d leaf entry\n", l.Name, len(entries.Entries))
			}

			for i := range entries.Entries {

				rawLogE, err := ct.RawLogEntryFromLeaf(int64(l.Index), &entries.Entries[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s -> Failed to parse leaf to raw entry at index %d: %s\n", l.Name, l.Index, err)
					l.Index++
					continue
				}

				l.Index++

				logE, err := rawLogE.ToLogEntry()
				if err != nil && logE == nil {
					fmt.Printf("%s -> Failed to convert raw log to log at index %d: %s\n", l.Name, rawLogE.Index, err)
					continue
				}

				if DEBUG_MODE {
					fmt.Printf("%s -> Leaf entry at %d is parsed successfuly!\n", l.Name, logE.Index)
				}

				// Fetch domains from cert and send it to writer through domainChan
				if logE.X509Cert != nil {

					if domain.IsValid(logE.X509Cert.Subject.CommonName) {
						domainChan <- strings.ToLower(logE.X509Cert.Subject.CommonName)
					}

					for i := range logE.X509Cert.DNSNames {
						if domain.IsValid(logE.X509Cert.DNSNames[i]) {
							domainChan <- strings.ToLower(logE.X509Cert.DNSNames[i])
						}
					}

					for i := range logE.X509Cert.PermittedDNSDomains {
						if domain.IsValid(logE.X509Cert.PermittedDNSDomains[i]) {
							domainChan <- strings.ToLower(logE.X509Cert.PermittedDNSDomains[i])
						}
					}
				}

				// Fetch domains from precert and send it to writer through domainChan
				if logE.Precert != nil && logE.Precert.TBSCertificate != nil {

					if domain.IsValid(logE.Precert.TBSCertificate.Subject.CommonName) {
						domainChan <- strings.ToLower(logE.Precert.TBSCertificate.Subject.CommonName)
					}

					for i := range logE.Precert.TBSCertificate.DNSNames {
						if domain.IsValid(logE.Precert.TBSCertificate.DNSNames[i]) {
							domainChan <- strings.ToLower(logE.Precert.TBSCertificate.DNSNames[i])
						}
					}

					for i := range logE.Precert.TBSCertificate.PermittedDNSDomains {
						if domain.IsValid(logE.Precert.TBSCertificate.PermittedDNSDomains[i]) {
							domainChan <- strings.ToLower(logE.Precert.TBSCertificate.PermittedDNSDomains[i])
						}
					}
				}
			}
		}
	}

	fmt.Printf("%s -> Finished parsing %d logs!\n", l.Name, l.Index)
}

func Fetch(dir string) error {

	defer fullFile.Close()
	defer subFile.Close()
	defer domainFile.Close()
	defer tldFile.Close()

	var err error

	fullFilePath = fmt.Sprintf("%s/columbus.full", dir)
	fullFile, err = os.OpenFile(fullFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", fullFilePath, err)
	}
	fmt.Printf("Path of the full hostnames: %s\n", fullFilePath)

	subFilePath = fmt.Sprintf("%s/columbus.sub", dir)
	subFile, err = os.OpenFile(subFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", fullFilePath, err)
	}
	fmt.Printf("Path of the subdomains: %s\n", subFilePath)

	domainFilePath = fmt.Sprintf("%s/columbus.domain", dir)
	domainFile, err = os.OpenFile(domainFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", fullFilePath, err)
	}
	fmt.Printf("Path of the domains: %s\n", domainFilePath)

	tldFilePath = fmt.Sprintf("%s/columbus.tld", dir)
	tldFile, err = os.OpenFile(tldFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", fullFilePath, err)
	}
	fmt.Printf("Path of the TLDs: %s\n", tldFilePath)

	domainChan = make(chan string, 512)

	if err := readIndex(dir); err != nil {
		return fmt.Errorf("failed to read index file: %s", err)
	}

	// Terminate context in case of SIGTERM
	sigCtx, sigShutdown := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer sigShutdown()

	// Start background saver
	go backgroundSave(sigCtx, dir)

	var writerWG sync.WaitGroup
	writerWG.Add(1)
	go writer(&writerWG)

	var wg sync.WaitGroup

	for i := range URIS {
		wg.Add(1)
		go fetcher(&wg, sigCtx, &URIS[i])
	}
	wg.Wait()

	// Close domainChan to terminate writer, the writerWG will wait until the last write
	close(domainChan)
	defer writerWG.Wait()

	if err := saveIndex(dir); err != nil {
		return fmt.Errorf("failed to save index: %s", err)
	}

	return nil
}
