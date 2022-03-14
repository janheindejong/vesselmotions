package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/websocket"
)

// https://github.com/gorilla/websocket/blob/69d0eb9187b6dead8fe84b2423518475e5cc535c/examples/chat/client.go

const (
	defaultBufferSize = 100
	writeWait         = 2 * time.Second
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
	clock clock.Clock
}

func (wsHandler *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// TODO subscription message

	client := Client{
		conn:          conn,
		hub:           wsHandler.hub,
		messageTicker: wsHandler.clock.Ticker(time.Duration(1e9)),
	}

	// Launch client that actually handles the sending of data
	go client.listenAndSend()
}

type Client struct {
	conn          *websocket.Conn
	hub           Subscribeable
	messageTicker *clock.Ticker
}

func (c *Client) listenAndSend() {
	buffer := c.hub.Subscribe(defaultBufferSize)
	defer func() {
		c.hub.UnSubscribe(buffer)
		c.conn.Close()
	}()
	points := []DataPoint{} // Create new empty slice of datapoints

	for {
		select {
		case p := <-buffer:
			// Add any incoming new points
			points = append(points, p)

		case <-c.messageTicker.C:
			// Empty list of points and send message
			msg := fmt.Sprintf(`NumberOfPoints="%v"`, len(points))
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return
			}

			// Empty list of datapoints
			points = []DataPoint{}
		}
	}
}
