package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elmasy-com/columbus/writer"
	"github.com/elmasy-com/slices"
	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
)

type log struct {
	URI    string            // URI of the log
	Name   string            // "nickname" of the log
	Client *client.LogClient // CT LogCLient
	index  int               // Index to start with, updated from the .index file if present
	size   int               // The number of entries  stored in the log
	ctx    context.Context   // TODO: Pointless
	toWait int
	err    error // Last error
	m      *sync.Mutex
}

var (
	USER_AGENT = "Elmasy-Columbus/0.1-dev"

	// From https://www.gstatic.com/ct/log_list/log_list.json
	URIS = []log{
		{"https://ct.googleapis.com/logs/argon2022/", "argon2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/logs/argon2023/", "argon2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/logs/xenon2022/", "xenon2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/logs/xenon2023/", "xenon2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/icarus/", "icarus", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/pilot/", "pilot", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/rocketeer/", "rocketeer", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.googleapis.com/skydiver/", "skydiver", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.cloudflare.com/logs/nimbus2022/", "nimbus2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.cloudflare.com/logs/nimbus2023/", "nimbus2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct1.digicert-ct.com/log/", "digicert-ct1", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct2.digicert-ct.com/log/", "digicert-ct2", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://yeti2022.ct.digicert.com/log/", "yeti2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://yeti2022-2.ct.digicert.com/log/", "yeti2022-2", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://yeti2023.ct.digicert.com/log/", "yeti2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://nessie2022.ct.digicert.com/log/", "nessie2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://nessie2023.ct.digicert.com/log/", "nessie2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://sabre.ct.comodo.com/", "sabre", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://mammoth.ct.comodo.com/", "mammoth", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://oak.ct.letsencrypt.org/2019/", "oak2019", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://oak.ct.letsencrypt.org/2020/", "oak2020", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://oak.ct.letsencrypt.org/2021/", "oak2021", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://oak.ct.letsencrypt.org/2022/", "oak2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://oak.ct.letsencrypt.org/2023/", "oak2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.trustasia.com/log2022/", "trustasia2022", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
		{"https://ct.trustasia.com/log2023/", "trustasia2023", nil, 0, 0, nil, 0, nil, new(sync.Mutex)},
	}

	STEP = 1000 // Number of log queried at once

	backgroundSaveInterval = 60   // Time to wait in second between two background save of indexes
	fetcherInterval        = 3600 // Time to wait in second between two successful fetch

	closer  context.CancelFunc
	running bool
	m       sync.Mutex
)

// Set index
func (l *log) setIndex(i int) {

	l.m.Lock()
	defer l.m.Unlock()

	l.index = i
}

// Increase the index by one
func (l *log) increaseIndex() {

	l.m.Lock()
	defer l.m.Unlock()

	l.index += 1
}

// Set the Error field and print is to STDERR.
// Format the string to "l.Name -> <error>\n"
func (l *log) setError(format string, a ...any) {

	l.m.Lock()
	defer l.m.Unlock()

	format = fmt.Sprintf("%s -> %s", l.Name, format)

	l.err = fmt.Errorf(format, a...)

	fmt.Fprintf(os.Stderr, "%s\n", l.err.Error())
}

// Update the size of the log from SignedTreeHead.
func (l *log) setSize() error {

	if l.Client == nil {
		return fmt.Errorf("client is nil")
	}

	l.m.Lock()
	defer l.m.Unlock()

	sth, err := l.Client.GetSTH(context.TODO())
	if err != nil {
		return err
	}

	l.size = int(sth.TreeSize)

	return nil
}

// Update the index of the log from the index file.
func (l *log) updateIndex() error {

	// Do not update log, that updated before
	if l.index > 0 {
		return nil
	}

	if _, err := os.Stat(writer.WorkingDir + "/index"); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	out, err := os.ReadFile(writer.WorkingDir + "/index")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")

	for i := range lines {

		fields := strings.Split(lines[i], "=")

		if len(fields) != 2 {
			continue
		}

		if fields[0] != l.Name {
			continue
		}

		ind, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index %s to int for %s", fields[1], fields[0])
		}

		l.setIndex(ind)
	}

	return nil
}

func (l *log) setCtx(ctx context.Context) {

	l.m.Lock()
	defer l.m.Unlock()

	l.ctx = ctx
}

// Set toWait to s seconds.
func (l *log) setToWait(s int) {

	l.m.Lock()
	defer l.m.Unlock()

	l.toWait = s
}

// Increase toWait seconds by 10
func (l *log) increaseToWait() {

	l.m.Lock()
	defer l.m.Unlock()

	l.toWait += 10
}

// Get raw entries from the log server, handle non fatal errors and return the raw entires.
func (l *log) GetRawEntries(start, end int64) ([]ct.LeafEntry, error) {

	entries, err := l.Client.GetRawEntries(context.TODO(), start, end)
	if err != nil {

		switch {
		case strings.Contains(err.Error(), "429 Too Many Requests"):
			l.increaseToWait()
			l.setError("Failed to get raw entries: %s. Waiting for %d seconds to retry...", err, l.toWait)
			return nil, nil
		case strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers"):
			// A temporary error, sleep 10 sec
			l.setError("Failed to get raw entries: %s. Waiting for 10 seconds to retry...", err)
			sleeper(l.ctx, 10)
			return nil, nil
		case strings.Contains(err.Error(), "Client.Timeout or context cancellation while reading body"):
			// Context is set to TODO() so, Client.Timeout must be the issue, retry after 10 sec
			l.setError("Failed to get raw entries: %s. Waiting for 10 seconds to retry...", err)
			sleeper(l.ctx, 10)
			return nil, nil
		default:
			return nil, err
		}
	}

	if entries == nil {
		return nil, fmt.Errorf("entries is nil")
	}

	return entries.Entries, nil
}

func getRunning() bool {
	m.Lock()
	defer m.Unlock()
	return running
}

// Set the internal running to v.
func setRunning(v bool) {
	m.Lock()
	defer m.Unlock()
	running = v
}

// sleeper is a context aware sleep function, which is sleep for s second or return if ctx is cancelled.
func sleeper(ctx context.Context, s int) {

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(s) * time.Second):
		return
	}
}

// Save log.Index.
func saveIndex() error {

	file, err := os.OpenFile(writer.WorkingDir+"/index", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := range URIS {

		URIS[i].m.Lock()

		// Skip logs, thats not working/no progress.
		if URIS[i].index == 0 {
			URIS[i].m.Unlock()
			continue
		}

		_, err = file.WriteString(fmt.Sprintf("%s=%d\n", URIS[i].Name, URIS[i].index))
		if err != nil {
			URIS[i].m.Unlock()
			return err
		}

		URIS[i].m.Unlock()
	}

	return nil
}

// backgroundSave saves the indexes of the URIS every <backgroundSaveInterval> seconds.
// This function is for the case of any unexpected termination.
// signal context will terminate this, but a last save will happen at the end of Fetch()
func backgroundSave(ctx context.Context) {

	for {

		select {
		case <-ctx.Done():
			fmt.Printf("Fetcher background saver closed!\n")
			return
		default:
			sleeper(ctx, backgroundSaveInterval)
			if err := saveIndex(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save indexes: %s\n", err)

			}
		}
	}
}

// fetch domains from LogEntry and send it to domainChan
func fetchDomain(e *ct.LogEntry) {

	// Use this with slices.Contain() to filter duplicated domains.
	domains := make([]string, 0)

	// Fetch domains from cert and send it to writer through domainChan
	if e.X509Cert != nil {

		if !slices.Contain(domains, e.X509Cert.Subject.CommonName) {
			domains = append(domains, e.X509Cert.Subject.CommonName)
		}

		for i := range e.X509Cert.DNSNames {
			if !slices.Contain(domains, e.X509Cert.DNSNames[i]) {
				domains = append(domains, e.X509Cert.DNSNames[i])
			}
		}

		for i := range e.X509Cert.PermittedDNSDomains {
			if !slices.Contain(domains, e.X509Cert.PermittedDNSDomains[i]) {
				domains = append(domains, e.X509Cert.PermittedDNSDomains[i])
			}
		}
	}

	// Fetch domains from precert and send it to writer through domainChan
	if e.Precert != nil && e.Precert.TBSCertificate != nil {

		if !slices.Contain(domains, e.Precert.TBSCertificate.Subject.CommonName) {
			domains = append(domains, e.Precert.TBSCertificate.Subject.CommonName)
		}

		for i := range e.Precert.TBSCertificate.DNSNames {
			if !slices.Contain(domains, e.Precert.TBSCertificate.DNSNames[i]) {
				domains = append(domains, e.Precert.TBSCertificate.DNSNames[i])
			}
		}

		for i := range e.Precert.TBSCertificate.PermittedDNSDomains {
			if !slices.Contain(domains, e.Precert.TBSCertificate.PermittedDNSDomains[i]) {
				domains = append(domains, e.Precert.TBSCertificate.PermittedDNSDomains[i])
			}
		}
	}

	// Write only unique domains
	for i := range domains {
		writer.Write(domains[i])
	}
}

// fetchLog updates the fileds of l log and start the fetchinf loop.
// In case of termination (ctx.Done()) this function exits.
func fetchLog(wg *sync.WaitGroup, ctx context.Context, l *log) {

	defer wg.Done()

	err := l.updateIndex()
	if err != nil {
		l.setError("FATAL: Failed to update index: %s", err)
		return
	}
	if l.index > 0 {
		fmt.Printf("%s -> Continuing from index %d\n", l.Name, l.index)
	}

	l.Client, err = client.New(
		l.URI,
		&http.Client{
			Timeout: 30 * time.Second,
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
		jsonclient.Options{UserAgent: USER_AGENT})

	if err != nil {
		l.setError("FATAL: Failed to create new client: %s", err)
		return
	}

	err = l.setSize()
	if err != nil {
		l.setError("FATAL: Failed to set size: %s", err)
		return
	}

	l.setCtx(ctx)

	fmt.Printf("%s -> Number of logs: %d\n", l.Name, l.size)

	// Number of seconds to wait.
	l.setToWait(0)

	// Logs MAY return fewer than the number of leaves requested. Only complete
	// if we actually got all the leaves we were expecting.
	// See more: https://github.com/google/certificate-transparency-go/blob/52d94d8cbab94d6698621839ab1a439d17ebbfb2/scanner/fetcher.go#L263
	for l.index < l.size {

		select {
		case <-ctx.Done():
			fmt.Printf("%s -> Fetcher closed!\n", l.Name)
			return
		default:

			// If context is done, this function return only after every single entries sent to the writer.
			sleeper(l.ctx, l.toWait)

			entries, err := l.GetRawEntries(int64(l.index), int64(l.index+STEP))
			if err != nil {
				l.setError("FATAL: Failed to get raw entries: %s", err)
				return
			}

			for i := range entries {

				rawLogE, err := ct.RawLogEntryFromLeaf(int64(l.index), &entries[i])
				if err != nil {
					l.setError("Failed to parse leaf to raw entry at index %d: %s", l.index, err)
					l.increaseIndex()
					continue
				}

				l.increaseIndex()

				logE, err := rawLogE.ToLogEntry()
				if err != nil && logE == nil {
					l.setError("Failed to convert raw log to log at index %d: %s", rawLogE.Index, err)
					continue
				}

				fetchDomain(logE)
			}
		}
	}

	fmt.Printf("%s -> Finished parsing %d logs!\n", l.Name, l.index)
}

func fetcher(ctx context.Context) {

	setRunning(true)

	// Start background saver
	go backgroundSave(ctx)

	var wg sync.WaitGroup

	for {

		select {
		case <-ctx.Done():
			if err := saveIndex(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save fetcher index: %s", err)
			}
			fmt.Printf("Indexes saved after cancelled fetcher...\n")
			setRunning(false)
			return
		default:

			// The fetchLog() goroutines also got the context, if the context is done,
			// this interation finishes and the new iteration trigers the ctx.Done()
			for i := range URIS {
				wg.Add(1)
				go fetchLog(&wg, ctx, &URIS[i])
			}
			wg.Wait()

			if err := saveIndex(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save fetcher index: %s", err)
			}
		}

		// Wait <fetcherInterval> seconds before starting a new fetching loop.
		sleeper(ctx, fetcherInterval)
	}
}

// Start starts fetcher in the background.
func Start() {

	// Do not start a new fetcher instance.
	if getRunning() {
		return
	}

	var ctx context.Context

	ctx, closer = context.WithCancel(context.Background())

	go fetcher(ctx)
}

func IsRunning() bool {
	return getRunning()
}

// Close blocks until
func Close() {

	if !getRunning() {
		return
	}

	closer()

	for getRunning() {
		time.Sleep(1 * time.Second)
	}
}
