package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func Serve(address string) {
	cert, err := genCertificate()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:      address,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		Handler:   http.HandlerFunc(new(Proxy).Serve),
	}

	log.Printf("listener serve at %s\n", address)
	log.Fatal(server.ListenAndServe())
}
