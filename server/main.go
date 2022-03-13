package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Parse arguments
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	// Setup hub
	hub := Hub{
		subscribers: map[chan DataPoint]string{},
	}

	// Setup publisher
	publisher := MockSensorDataPublisher{
		hub: &hub,
	}

	// Setup websocket handler
	wsHandler := NewWebSocketHandler(&hub)

	// Setup endpoints
	s := http.NewServeMux()
	s.Handle("/", wsHandler)

	// Run stuff
	go publisher.RunForever()
	log.Print("Serving on: ", *addr)
	log.Fatal(http.ListenAndServe(*addr, s))
}
