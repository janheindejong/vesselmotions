package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWsHandl(t *testing.T) {
	// Prepare
	hub := Hub{
		subscribers: map[chan DataPoint]string{},
	}
	wsHandler := WebSocketHandler{
		hub:      &hub,
		upgrader: websocket.Upgrader{},
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
		hub.Broadcast(point)
	}

	_, p, err := ws.ReadMessage() // TODO readJSON
	if err != nil {
		t.Fatalf("Could not read message: %v", err)
	}

	// Assert
	got := string(p)
	want := `NumberOfPoints="2"`
	if got != want {
		t.Errorf("Incorrect message, got: %v, want: %v", got, want)
	}

}
