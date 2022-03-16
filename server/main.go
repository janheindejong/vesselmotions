package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/benbjohnson/clock"
)

func main() {
	// Parse arguments
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	// Setup hub
	hub := Hub[DataPoint]{
		subscribers: map[chan DataPoint]string{},
	}

	// Setup publisher
	publisher := MockSensorDataPublisher{
		hub: &hub,
	}

	// Setup websocket handler
	wsHandler := WebSocketHandler{
		hub:   &hub,
		clock: clock.New(),
	}

	// Setup endpoints
	s := http.NewServeMux()
	s.Handle("/", &wsHandler)

	// Run stuff
	go publisher.RunForever()
	log.Print("Serving on: ", *addr)
	log.Fatal(http.ListenAndServe(*addr, s))
}
