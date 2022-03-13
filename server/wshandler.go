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

	client := Client{
		conn: conn,
		hub:  wsHandler.hub,
	}

	go client.listenAndSend()
}

type Client struct {
	conn *websocket.Conn
	hub  Subscribeable
}

func (c *Client) listenAndSend() {
	buffer := c.hub.Subscribe(defaultBufferSize)
	defer func() {
		c.hub.UnSubscribe(buffer)
		c.conn.Close()
	}()
	pingTicker := time.NewTicker(time.Duration(1e6))
	messageTicker := time.NewTicker(time.Duration(1e9))
	points := []DataPoint{} // Create new empty slice of datapoints

	for {
		select {
		case p := <-buffer:
			points = append(points, p)
		case <-messageTicker.C:
			// Empty list of points and send message
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			msg := fmt.Sprintf(`NumberOfPoints="%v"`, len(points))
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return
			}
			points = []DataPoint{}
		case <-pingTicker.C:
			// Ping, and close if stuff is broken
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
