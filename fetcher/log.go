package fetcher

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elmasy-com/columbus/writer"
	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
)

type log struct {
	URI    string            // URI of the log
	name   string            // "nickname" of the log
	client *client.LogClient // CT LogCLient
	index  int               // Index to start with, updated from the index file if present
	size   int               // The number of entries  stored in the log
	ctx    context.Context   // TODO: Pointless
	toWait int
	err    error // Last error
	m      *sync.Mutex
}

func (l *log) GetURI() string {
	l.m.Lock()
	defer l.m.Unlock()

	return l.URI
}

func (l *log) GetName() string {

	l.m.Lock()
	defer l.m.Unlock()

	return l.name
}

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

func (l *log) GetIndex() int {

	l.m.Lock()
	defer l.m.Unlock()

	return l.index
}

// Set the Error field and print is to STDERR.
// Format the string to "l.Name -> <error>\n"
func (l *log) setError(format string, a ...any) {

	l.m.Lock()
	defer l.m.Unlock()

	format = fmt.Sprintf("%s -> %s", l.name, format)

	l.err = fmt.Errorf(format, a...)

	fmt.Fprintf(os.Stderr, "%s\n", l.err.Error())
}

func (l *log) GetError() error {

	l.m.Lock()
	defer l.m.Unlock()

	return l.err
}

func (l *log) GetSize() int {

	l.m.Lock()
	defer l.m.Unlock()

	return l.size
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

// Update the index of the log from the index file.
func (l *log) updateIndex() error {

	l.m.Lock()
	defer l.m.Unlock()

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

		if fields[0] != l.name {
			continue
		}

		ind, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index %s to int for %s", fields[1], fields[0])
		}

		l.index = ind
	}

	return nil
}

// Set the context for the log. THis context is same as in every other log.
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

	entries, err := l.client.GetRawEntries(context.TODO(), start, end)
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
