package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// PubSub
var hub = Hub{
	subscribers: map[chan DataPoint]string{},
}
var publisher = MockPublisher{
	hub: &hub,
}

// Websockets
var upgrader = websocket.Upgrader{}

// Command line args
var addr = flag.String("addr", "localhost:8080", "http service address")

func sensorDataHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Create buffer
	buffer := make(chan DataPoint, 100)
	hub.Subscribe(buffer)
	defer hub.UnSubscribe(buffer)

	// Publish messages
	for {
		points := emptyBuffer(buffer)
		msg := createMessage(points)
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Print(err)
			break
		}
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
	log.Println("Closing connection: ", err)
}

func createMessage(points *[]DataPoint) string {
	return fmt.Sprintf(`NumberOfPoints="%v"`, len(*points))
}

func main() {
	// Parse arguments
	flag.Parse()

	log.SetFlags(0)

	// Setup endpoints
	http.HandleFunc("/sensordata", sensorDataHandler)

	// Run stuff
	go publisher.RunForever()
	log.Print("Serving on: ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
