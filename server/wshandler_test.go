package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// MockHub is used for testing, to be able to access the channel used by the
// WebSocketHandler directly, and verify the calls to Subscribe and UnSubscribe
type mockHub struct {
	Channel    chan DataPoint
	Subscribed int
}

func (h *mockHub) Subscribe(bufferSize int) chan DataPoint {
	h.Subscribed++
	return h.Channel
}

func (h *mockHub) UnSubscribe(channel chan DataPoint) {
	h.Subscribed--
}

func TestWsHandler(t *testing.T) {
	// Prepare
	timeout := make(chan time.Time)
	hub := mockHub{
		Channel:    make(chan DataPoint),
		Subscribed: 0,
	}
	wsHandler := WebSocketHandler{
		hub:   &hub,
		after: func(d time.Duration) <-chan time.Time { return timeout },
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
		hub.Channel <- point
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
	t.Run("NumberOfPointsReceived", func(t *testing.T) {
		got := string(p)
		want := `NumberOfPoints="2"`
		if got != want {
			t.Errorf("Incorrect message, got: %v, want: %v", got, want)
		}
	})

	// time.Sleep(time.Duration(10) * time.Second)
	// if hub.Subscribed != 0 {
	// 	t.Error("Did not subscribe succesfully")
	// }

}
