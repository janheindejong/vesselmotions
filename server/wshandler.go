package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{}
)

type Subscribeable interface {
	Subscribe(int) chan DataPoint
	UnSubscribe(chan DataPoint)
}

type WebSocketHandler struct {
	hub   Subscribeable
	after func(time.Duration) <-chan time.Time
}

func (wsHandler *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()

	// Create and subscribe buffer
	buffer := wsHandler.hub.Subscribe(100)
	defer wsHandler.hub.UnSubscribe(buffer)

	// Start publishing messages at regular intervals
	for {
		points := wsHandler.listenForPoints(buffer, time.Duration(1e9))
		msg := wsHandler.fmtMessage(points)
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("closing connection: ", err)
			break // Break outer loop
		}
	}
}

func (wsHandler *WebSocketHandler) listenForPoints(buffer chan DataPoint, d time.Duration) *[]DataPoint {
	timeout := wsHandler.after(d) // Set new timeout
	points := []DataPoint{}       // Create new empty slice of datapoints
	// Listen to points coming on channel until timeout
	for {
		select {
		case point := <-buffer:
			points = append(points, point)
		case <-timeout:
			return &points
		}
	}
}

func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub:   hub,
		after: time.After,
	}
}

func (wsHandler *WebSocketHandler) fmtMessage(points *[]DataPoint) string {
	return fmt.Sprintf(`NumberOfPoints="%v"`, len(*points))
}
