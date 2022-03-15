package main

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/websocket"
)

const (
	timeLayout = "2006-01-02T15:04:05-07:00"
)

// MockHub is used for testing, to be able to access the channel used by the
// WebSocketHandler directly, and verify the calls to Subscribe and UnSubscribe
type mockedHub struct {
	Channel    chan DataPoint
	Subscribed int
}

func (h *mockedHub) Subscribe(bufferSize int) chan DataPoint {
	h.Subscribed++
	return h.Channel
}

func (h *mockedHub) UnSubscribe(channel chan DataPoint) {
	h.Subscribed--
}

func TestWsHandler(t *testing.T) {
	// Prepare

	// Mocked Clock
	mockedClock := clock.NewMock()
	now, _ := time.Parse(timeLayout, "1987-10-22T00:00:00+00:00")
	mockedClock.Set(now)

	// Mocked Hub
	mockedHub := mockedHub{
		Channel:    make(chan DataPoint),
		Subscribed: 0,
	}

	// WebSocketHandler and test server
	wsHandler := WebSocketHandler{
		hub:   &mockedHub,
		clock: mockedClock,
	}
	server := httptest.NewServer(&wsHandler)
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Datapoints - timestamps to UTC for comparisons
	points := []DataPoint{
		{
			Value:     0,
			Timestamp: now.UTC(),
			Id:        "",
		},
		{
			Value:     1,
			Timestamp: now.Add(time.Duration(500 * time.Millisecond)).UTC(),
			Id:        "",
		},
	}

	// Act

	// Open websocket connection to test server
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}
	defer ws.Close()

	// Send points to blocking channel, to ensure client takes them out
	// of the queue before proceeding
	for _, p := range points {
		mockedHub.Channel <- p
	}
	// Forward time by 1 second, to trigger message send
	mockedClock.Add(time.Duration(1) * time.Second)

	// Read and parse message
	_, data, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	var msg Message
	err = json.Unmarshal(data, &msg)
	if err != nil {
		t.Fatalf("could not parse message: %v", err)
	}

	// Close websocket from client side, to trigger teardown on server
	err = ws.Close()
	if err != nil {
		t.Fatalf("error while closing websocket: %v", err)
	}
	// Increment time by 2 seconds, to trigger 2 messages
	// First message doesn't trigger an error for some reason
	mockedClock.Add(time.Duration(2) * time.Second)

	// Assert
	t.Run("DataPointsReceived", func(t *testing.T) {
		got := msg.Data
		want := points
		if !reflect.DeepEqual(got, want) {
			t.Errorf("incorrect data, got: %v, want: %v", got, want)
		}
	})

	t.Run("MessageTime", func(t *testing.T) {
		got := msg.Timestamp
		want := now.Add(time.Duration(1) * time.Second).UTC()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("incorrect message time, got: %v, want: %v", got, want)
		}
	})

	t.Run("UnSubscribe", func(t *testing.T) {
		if mockedHub.Subscribed != 0 {
			t.Error("did not subscribe client succesfully")
		}
	})
}
