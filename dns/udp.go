package main

import (
	"fmt"
	"os"

	"github.com/miekg/dns"
)

// UDPStart starts the UDP server.
// If any error occured in ListenAndServe(), sends an os.Interupt into stopSignal.
func UDPStart(listen string, stopSignal chan os.Signal) *dns.Server {

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", handleFunc)

	udpServer := &dns.Server{
		Addr:    listen,
		Net:     "udp",
		Handler: udpHandler,
	}

	go func() {
		fmt.Printf("Starting UDP server...\n")
		err := udpServer.ListenAndServe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "UDP server failed: %s\n", err)
			stopSignal <- os.Interrupt
		}
	}()

	return udpServer
}
