package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/elmasy-com/columbus/db"
	eldns "github.com/elmasy-com/elnet/dns"

	"github.com/miekg/dns"
)

var (
	ReplyChan    chan *dns.Msg
	BuildDate    string
	BuildCommit  string
	resolvers    []string
	resolversNum int32
)

func getRandomResolver() string {

	if resolversNum == 1 {
		return resolvers[0]
	}

	return resolvers[rand.Int31n(resolversNum)]
}

// isValidResponse checks the type and the content of m.
// If m indicates a valid reply, returns true.
// This function is needed to not rely on RCODE only.
func isValidResponse(m dns.RR) bool {

	switch t := m.(type) {

	case *dns.A:
		return true
	case *dns.AAAA:
		return true
	case *dns.CAA:
		return true
	case *dns.CNAME:
		return true
	case *dns.DNAME:
		return true
	case *dns.MX:
		return true
	case *dns.NS:
		return true
	case *dns.SOA:
		return true
	case *dns.SRV:
		return true
	case *dns.TXT:
		return true

	//case *dns.CERT:
	// TODO: Implement more type
	//return false

	case *dns.PTR:
		// PTR records are out of context
		return false
	default:
		fmt.Fprintf(os.Stderr, "Unknown reply type: %T\n", t)
		return false
	}
}

// insertWorker is a goroutine.
// NumWorkers controls the number of workers.
func insertWorker(wg *sync.WaitGroup) {

	defer wg.Done()

	for r := range ReplyChan {

		switch {
		case len(r.Question) < 1:
			continue
		case len(r.Answer) < 1:
			continue
		case len(r.Question) > 1:
			continue
		case !isValidResponse(r.Answer[0]):
			// Further check requires, see: https://community.cloudflare.com/t/noerror-response-for-not-exist-domain-breaks-nslookup/173897
			continue
		}

		wc, err := eldns.IsWildcard(r.Question[0].Name, r.Question[0].Qtype)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check if %s is wildcard: %s\n", r.Question[0].Name, err)
			continue
		}
		if wc {
			continue
		}

		ni, err := db.DomainsInsert(r.Question[0].Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to insert %s: %s\n", r.Question[0].Name, err)
			continue
		}
		if ni {
			fmt.Printf("New domain inserted: %s\n", r.Question[0].Name)
		}

		err = db.RecordsUpdate(r.Question[0].Name, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update records for %s: %s\n", r.Question[0].Name, err)
		}
	}
}

// Return whether r.Question is ANY type.
func isQuestionAny(r *dns.Msg) bool {

	// Is ANY?
	for i := range r.Question {
		if r.Question[i].Qtype == dns.TypeANY {
			return true
		}
	}

	return false
}

func handleFunc(w dns.ResponseWriter, q *dns.Msg) {

	start := time.Now()

	// Refuse ANY questions
	if isQuestionAny(q) {
		q.MsgHdr.Response = true
		q.MsgHdr.Rcode = dns.RcodeNotImplemented
		w.WriteMsg(q)
		return
	}

	r, err := dns.Exchange(q, getRandomResolver())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to exchange message: %s\n", err)
		q.MsgHdr.Response = true
		q.MsgHdr.Rcode = dns.RcodeServerFailure
		w.WriteMsg(q)
		return
	}
	if r == nil {
		fmt.Fprintf(os.Stderr, "Error: reply is nil\n")
		q.MsgHdr.Response = true
		q.MsgHdr.Rcode = dns.RcodeServerFailure
		w.WriteMsg(q)
		return
	}

	if r.Rcode == dns.RcodeSuccess && len(ReplyChan) < cap(ReplyChan) {
		ReplyChan <- r
	}

	err = w.WriteMsg(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write reply: %s\n", err)
	}

	fmt.Printf("%s -> %s %s %s %s %s\n",
		w.RemoteAddr().String(),
		q.Question[0].Name,
		dns.ClassToString[q.Question[0].Qclass],
		dns.TypeToString[q.Question[0].Qtype],
		dns.RcodeToString[r.Rcode],
		time.Since(start))

}

func main() {

	configPath := flag.String("config", "", "Path to the config file")
	printVersion := flag.Bool("version", false, "Print version")
	flag.Parse()

	// Print version and  exit
	if *printVersion {
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", BuildCommit)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "-config is empty!\n")
		fmt.Printf("Use -help for help\n")
		os.Exit(1)
	}

	conf, err := parseConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config file: %s\n", err)
		os.Exit(1)
	}

	// Set global resolvers to get random ones
	resolvers = conf.Resolvers
	resolversNum = int32(len(resolvers))

	// Create buff channel
	ReplyChan = make(chan *dns.Msg, conf.BuffSize)

	// Connect to MongoDB
	fmt.Printf("Connecting to MongoDB...\n")
	err = db.Connect(conf.MongoURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to MongoDB: %s\n", err)
		os.Exit(1)
	}
	defer db.Disconnect()

	fmt.Printf("Starting %d workers...\n", conf.NumWorkers)
	// Start workers
	wg := sync.WaitGroup{}
	for i := 0; i < conf.NumWorkers; i++ {
		wg.Add(1)
		go insertWorker(&wg)
	}

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	udpServer := UDPStart(conf.ListenAddress, stopSignal)

	tcpServer := TCPStart(conf.ListenAddress, stopSignal)

	// Wait for the SIGTERM
	<-stopSignal
	fmt.Printf("Caught a SIGTERM, closing...\n")
	udpServer.Shutdown()
	tcpServer.Shutdown()
	close(ReplyChan)
	wg.Wait()
}
