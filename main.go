package main

import (
	"flag"
	"log"
)

var (
	address string
)

func init() {
	flag.StringVar(&address, "l", "0.0.0.0:8080", "listener address e.g: 127.0.0.1:8080")
}

func main() {
	flag.Parse()

	if !checkAddress(address) {
		log.Fatalf("the listener address [%s] incorrect, please check it", address)
	}
	Serve(address)
}
