package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub      *Hub
	upgrader websocket.Upgrader
}

func (handler *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade websocket
	c, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Create buffer
	buffer := make(chan DataPoint, 100)
	handler.hub.Subscribe(buffer)
	defer handler.hub.UnSubscribe(buffer)

	// Publish messages
	for {
		points := emptyBuffer(buffer)
		msg := createMessage(points)
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("closing connection: ", err)
			break
		}
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
}

func createMessage(points *[]DataPoint) string {
	return fmt.Sprintf(`NumberOfPoints="%v"`, len(*points))
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
