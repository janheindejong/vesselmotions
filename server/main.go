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
	subscribers: map[Subscriber]string{},
}
var publisher = MockPublisher{
	hub: &hub,
}

// Websockets
var upgrader = websocket.Upgrader{}

// Command line args
var addr = flag.String("addr", "localhost:8080", "http service address")

func SensorData(w http.ResponseWriter, r *http.Request) {
	// Upgrade websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Create client
	client := Client{
		channel:            make(chan DataPoint, 100),
		publishingInterval: 1000,
	}
	hub.Subscribe(&client)
	defer hub.UnSubscribe(&client)

	// Publish messages
	err = client.Run(c)
	log.Println("Closed client: ", err)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	// Setup endpoints
	http.HandleFunc("/sensordata", SensorData)

	// Run stuff
	go publisher.Run()
	log.Print("Serving on: ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type Client struct {
	channel            chan DataPoint
	publishingInterval int
}

func (client *Client) Channel() chan DataPoint {
	return client.channel
}

func (client *Client) Run(c *websocket.Conn) error {
	log.Print("Running client")
	for {
		points := emptyBuffer(client.channel)
		msg := createMessage(points)
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Print(err)
			return err
		}
		time.Sleep(time.Millisecond * time.Duration(client.publishingInterval))
	}
}

func emptyBuffer(channel chan DataPoint) *[]DataPoint {
	datapoints := make([]DataPoint, 0)
	for {
		select {
		case NewDataPoint := <-channel:
			datapoints = append(datapoints, NewDataPoint)
		default:
			return &datapoints
		}
	}
}

func createMessage(points *[]DataPoint) string {
	return fmt.Sprintf(`NumberOfPoints="%v"`, len(*points))
}

// Receiver gets raw data; for now it's just a producer of random stuff

// Hub is where you subscribe

// Client is a consumer of the hub

// Main will handle creation of clients and the webscokets server
