package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// BlockedBroadcast is a way to make sure the test only continues after all
// datapoints have been consumed by the websocket goroutine
func (hub *Hub) BlockedBroadcast(point DataPoint) {
	for subscriber := range hub.subscribers {
		subscriber <- point
	}
}

func TestWsHandler(t *testing.T) {
	// Prepare
	timeout := make(chan time.Time)
	hub := Hub{
		subscribers: map[chan DataPoint]string{},
	}
	wsHandler := WebSocketHandler{
		hub:      &hub,
		upgrader: websocket.Upgrader{},
		after:    func(d time.Duration) <-chan time.Time { return timeout },
		bufferFactory: func(size int) chan DataPoint {
			return make(chan DataPoint)
		},
	}
	server := httptest.NewServer(&wsHandler)
	defer server.Close()

	wsUrl := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	datapoints := []DataPoint{
		{
			Value:     0,
			Timestamp: time.Time{},
			Id:        "",
		},
		{
			Value:     1,
			Timestamp: time.Time{},
			Id:        "",
		},
	}

	// Act
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsUrl, err)
	}
	defer ws.Close()

	for _, point := range datapoints {
		hub.BlockedBroadcast(point)
	}
	timeout <- time.Time{}

	_, p, err := ws.ReadMessage() // TODO readJSON
	if err != nil {
		t.Fatalf("Could not read message: %v", err)
	}

	err = ws.Close()
	if err != nil {
		t.Fatalf("Error while closing websocket: %v", err)
	}
	timeout <- time.Time{}

	// Assert
	got := string(p)
	want := `NumberOfPoints="2"`
	if got != want {
		t.Errorf("Incorrect message, got: %v, want: %v", got, want)
	}

	if len(hub.subscribers) != 0 {
		t.Error("Did not subsribe succesfully")
	}

}
