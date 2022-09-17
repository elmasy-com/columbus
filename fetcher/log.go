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

	"github.com/elmasy-com/columbus/config"
	"github.com/elmasy-com/columbus/writer"
	"github.com/elmasy-com/slices"
	"github.com/g0rbe/slitu"
	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
)

type log struct {
	uri            string            // URI of the log
	name           string            // "nickname" of the log
	index          int               // Index to start with, updated from the index file if present
	size           int               // The number of entries  stored in the log
	wait           int               // Time ti wait in seconds between to fetch
	err            error             // Last error
	client         *client.LogClient // CT LogCLient
	fetcherRunning bool              // Indicate that fetcher is running
	indexFilePath  string
	ctx            context.Context
	cancel         context.CancelFunc
	m              sync.Mutex
}

/*
 * Getters
 */

func (l *log) GetURI() string {
	l.m.Lock()
	defer l.m.Unlock()

	return l.uri
}

func (l *log) GetName() string {

	l.m.Lock()
	defer l.m.Unlock()

	return l.name
}

func (l *log) GetIndex() int {

	l.m.Lock()
	defer l.m.Unlock()

	return l.index
}

func (l *log) GetSize() int {

	l.m.Lock()
	defer l.m.Unlock()

	return l.size
}

func (l *log) getWait() int {

	l.m.Lock()
	defer l.m.Unlock()

	return l.wait
}

func (l *log) GetError() error {

	l.m.Lock()
	defer l.m.Unlock()

	return l.err
}

// GetRunning returns that both fetcher and writer is running.
// Returns false if one of them are stopped.
func (l *log) GetRunning() bool {

	l.m.Lock()
	defer l.m.Unlock()

	return l.fetcherRunning
}

func (l *log) getFetcherRunning() bool {

	l.m.Lock()
	defer l.m.Unlock()

	return l.fetcherRunning
}

/*
 * Setters
 */

// Set the error field and print is to STDERR.
// Format the string to "l.Name -> <error>\n"
func (l *log) setError(format string, a ...any) {

	l.m.Lock()
	defer l.m.Unlock()

	l.err = fmt.Errorf(format, a...)

	fmt.Fprintf(os.Stderr, "%s -> %s\n", l.name, l.err.Error())
}

// Update the size of the log from SignedTreeHead.
func (l *log) setSize() error {

	if l.client == nil {
		return fmt.Errorf("client is nil")
	}

	l.m.Lock()
	defer l.m.Unlock()

	sth, err := l.client.GetSTH(context.TODO())
	if err != nil {
		return err
	}

	l.size = int(sth.TreeSize)

	return nil
}

func (l *log) setFetcherRunning(v bool) {

	l.m.Lock()
	defer l.m.Unlock()

	l.fetcherRunning = v
}

/*
 * Others
 */

// Increase the index by one
func (l *log) increaseIndex() {

	l.m.Lock()
	defer l.m.Unlock()

	l.index += 1
}

// Update the index of the log from the index file.
func (l *log) updateIndex() error {

	l.m.Lock()
	defer l.m.Unlock()

	// Do not update log, that updated before
	if l.index > 0 {
		return nil
	}

	out, err := os.ReadFile(l.indexFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Removes the \n in case if someone mess with the file (eg.: nano loves to append \n)
	l.index, err = strconv.Atoi(strings.Trim(string(out), "\n"))
	if err != nil {
		return fmt.Errorf("failed to convert %s to int: %s", out, err)
	}

	return nil
}

// Increase toWait seconds by 10
func (l *log) increaseToWait() {

	l.m.Lock()
	defer l.m.Unlock()

	l.wait += 10
}

// Save the last parsed index number.
func (l *log) saveIndex() error {

	// Do not save 0 as index
	if l.GetIndex() == 0 {
		return nil
	}

	file, err := os.OpenFile(l.indexFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d", l.GetIndex()))
	if err != nil {
		return err
	}

	return nil
}

// backgroundSave saves the indexes of the URIS every <backgroundSaveInterval> seconds.
// This function is for the case of any unexpected termination.
// signal context will terminate this, but a last save will happen at the end of Fetch()
func (l *log) backgroundSave() {

	// TODO: What if finished parsing logs?

	var err error

	for {

		select {
		case <-l.ctx.Done():
			fmt.Printf("%s -> Background saver closed!\n", l.name)
			return
		default:

			// Log parsed, wait until the new fetch iteration
			if l.GetIndex() >= l.GetSize() {
				slitu.Sleep(l.ctx, time.Duration(config.FetcherInterval)*time.Second)
				continue
			}

			slitu.Sleep(l.ctx, time.Duration(config.BackgroundSaveInterval)*time.Second)

			if err = l.saveIndex(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save indexes: %s\n", err)
			}
		}
	}
}

/*
 * Log related
 */

// parse unique domains from LogEntry and send it to domains channel.
func (l *log) fetchDomains(e *ct.LogEntry) {

	// Use this with slices.Contain() to filter duplicated domains (prefilter).
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

	// Write only unique and valid domains
	for i := range domains {
		writer.Write(domains[i])
	}
}

func (l *log) fetch() {

	l.setFetcherRunning(true)
	defer l.setFetcherRunning(false)
	// Save the last index at the end but dont handle error (!!)
	defer l.saveIndex()

	// "Infinity loop" that only l.ctx can stop
	// it fetch -> unique -> sleep -> get new STH
	//      '------------------------------'
	for {

		//toBeParsed is used to calculate how many logs are remaining
		toBeParsed := l.GetSize() - l.GetIndex()

		/*
		 * Fetching loop
		 */
		for l.GetIndex() < l.GetSize() {

			select {
			case <-l.ctx.Done():
				fmt.Printf("%s -> Fetcher closed!\n", l.GetName())
				return
			default:

				// TODO: index + step can overflow (can it be an issue?)
				entries, err := l.client.GetRawEntries(context.TODO(), int64(l.GetIndex()), int64(l.GetIndex()+config.Step))
				if err != nil {

					switch {
					case strings.Contains(err.Error(), "429 Too Many Requests"):
						// Too many request, nothing to see here. Wait and restart fetching loop.
						l.increaseToWait()
						//l.setError("Failed to get raw entries: %s. Waiting for %d seconds to retry...", err, l.getWait())
						slitu.Sleep(l.ctx, time.Duration(l.getWait())*time.Second)
						continue
					case strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers"):
						// A temporary error, sleep 10 sec and restart fetching loop
						//l.setError("Failed to get raw entries: %s. Waiting for 10 seconds to retry...", err)
						slitu.Sleep(l.ctx, 10*time.Second)
						continue
					case strings.Contains(err.Error(), "Client.Timeout or context cancellation while reading body"):
						// Context is set to TODO() so, Client.Timeout must be the issue, restart fetching loop after 10 sec
						l.setError("Failed to get raw entries: %s. Waiting for 10 seconds to retry...", err)
						slitu.Sleep(l.ctx, time.Duration(l.getWait())*time.Second)
						continue
					default:
						l.setError("Failed to get raw entries: %s", err)
						return
					}
				}

				if entries == nil {
					l.setError("entries is nil")
					return
				}

				// Convert raw log entry to log entry one by one.
				for i := range entries.Entries {

					rawLogE, err := ct.RawLogEntryFromLeaf(int64(l.GetIndex()), &entries.Entries[i])
					if err != nil {
						l.setError("Failed to parse leaf to raw entry at index %d: %s", l.GetIndex(), err)
						l.increaseIndex()
						continue
					}

					l.increaseIndex()

					logE, err := rawLogE.ToLogEntry()
					if err != nil && logE == nil {
						l.setError("Failed to convert raw log to log at index %d: %s", rawLogE.Index, err)
						continue
					}

					l.fetchDomains(logE)
				}
			}
		}

		/*
		 * Afterwork:
		 *		Makes the entries in the list file unique by stopping writer, call unique() and restart writer.
		 */
		select {
		case <-l.ctx.Done():
			fmt.Printf("%s -> Fetcher closed!\n", l.GetName())
			return
		default:

			/*
			 * In case of ctx.Done() the inifity loop start the "Fetching loop" (if new entry present) and stops the fetcher from there
			 * or go to afterwork's case and stops from there.
			 */

			// Do not make unique and restart writer if nothing happend
			if toBeParsed != 0 {
				fmt.Printf("%s -> Finished parsing %d logs!\n", l.GetName(), toBeParsed)
			}

			slitu.Sleep(l.ctx, time.Duration(config.FetcherInterval)*time.Second)

			err := l.setSize()
			if err != nil {
				l.setError("failed to set size in fetcher: %s\n", err)
				return
			}
		}
	}
}

// Close the log.
// This function blocks until everything is closed.
func (l *log) Close() {

	// Stop fetcher only if running
	if l.getFetcherRunning() {

		l.cancel()

		for l.getFetcherRunning() {
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// Fils the fields of log l and start background saver, writer and fetcher.
func (l *log) Start() {

	var err error

	l.client, err = client.New(
		l.uri,
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
		jsonclient.Options{UserAgent: config.UserAgent})

	if err != nil {
		l.setError("Failed to create client: %s", err)
		return
	}

	l.indexFilePath = fmt.Sprintf("%s/%s.index", config.WorkingDir, l.name)

	err = l.updateIndex()
	if err != nil {
		l.setError("Failed to update index: %s", err)
		return
	}

	if l.index > 0 {
		fmt.Printf("%s -> Continue from index %d\n", l.name, l.index)
	}

	err = l.setSize()
	if err != nil {
		l.setError("Failed to set size: %s", err)
		return
	}

	fmt.Printf("%s -> Number of logs: %d\n", l.GetName(), l.GetSize())

	l.ctx, l.cancel = context.WithCancel(context.Background())

	go l.backgroundSave()
	go l.fetch()
}
