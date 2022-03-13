package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Subscribeable interface {
	Subscribe(chan DataPoint)
	UnSubscribe(chan DataPoint)
}

type WebSocketHandler struct {
	hub           Subscribeable
	upgrader      websocket.Upgrader
	after         func(time.Duration) <-chan time.Time
	bufferFactory func(size int) chan DataPoint
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
	buffer := handler.bufferFactory(100)
	handler.hub.Subscribe(buffer)
	defer handler.hub.UnSubscribe(buffer)

	// Publish messages at regular intervals
	for {
		timeout := handler.after(time.Duration(1000) * time.Millisecond) // Set new timeout
		points := []DataPoint{}                                          // Create new empty slice of datapoints
		// Listen to points coming on channel until timeout
	getdata:
		for {
			select {
			case point := <-buffer:
				points = append(points, point)
			case <-timeout:
				break getdata
			}
		}
		// Send message
		msg := createMessage(&points)
		// c.SetWriteDeadline(time.Now().Add(time.Duration(10) * time.Millisecond))
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("closing connection: ", err)
			break // Break outer loop

		}
	}
}

func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub:      hub,
		upgrader: websocket.Upgrader{},
		after:    time.After,
		bufferFactory: func(size int) chan DataPoint {
			return make(chan DataPoint, size)
		},
	}
}

func createMessage(points *[]DataPoint) string {
	return fmt.Sprintf(`NumberOfPoints="%v"`, len(*points))
}
