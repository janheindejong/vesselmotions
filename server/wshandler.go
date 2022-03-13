package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

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
	hub Subscribeable
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
		pingTicker:    time.NewTicker(time.Duration(1e6)),
		messageTicker: time.NewTicker(time.Duration(1e9)),
	}

	go client.listenAndSend()
}

type Client struct {
	conn          *websocket.Conn
	hub           Subscribeable
	pingTicker    *time.Ticker
	messageTicker *time.Ticker
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			msg := fmt.Sprintf(`NumberOfPoints="%v"`, len(points))
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return
			}

			// Empty list of datapoints
			points = []DataPoint{}

		case <-c.pingTicker.C:
			// Ping, and close if stuff is broken
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
