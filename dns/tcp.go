package main

import (
	"fmt"
	"os"

	"github.com/miekg/dns"
)

// TCPStart starts the TCP server.
// If any error occured in ListenAndServe(), sends an os.Interupt into stopSignal.
func TCPStart(listen string, stopSignal chan os.Signal) *dns.Server {

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", handleFunc)

	tcpServer := &dns.Server{
		Addr:    listen,
		Net:     "tcp",
		Handler: tcpHandler,
	}

	go func() {
		fmt.Printf("Starting TCP server...\n")
		err := tcpServer.ListenAndServe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "TCP server failed: %s\n", err)
			stopSignal <- os.Interrupt
		}
	}()

	return tcpServer
}
