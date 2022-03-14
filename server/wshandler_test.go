package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/websocket"
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

const (
	timeLayout = "2006-01-02T15:04:05-07:00"
)

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

	// WebSocketHandler and server
	wsHandler := WebSocketHandler{
		hub:   &mockedHub,
		clock: mockedClock,
	}
	server := httptest.NewServer(&wsHandler)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Datapoints
	points := []DataPoint{
		{
			Value:     0,
			Timestamp: now,
			Id:        "",
		},
		{
			Value:     1,
			Timestamp: now.Add(time.Duration(500 * time.Millisecond)),
			Id:        "",
		},
	}

	// Act
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}
	defer ws.Close()

	// Send points to blocking channel, to ensure client takes them out of the queue
	for _, p := range points {
		mockedHub.Channel <- p
	}
	// Forward time by 1 second, to trigger message send
	mockedClock.Add(time.Duration(1) * time.Second)

	_, data, err := ws.ReadMessage() // TODO readJSON
	if err != nil {
		t.Fatalf("Could not read message: %v", err)
	}

	err = ws.Close()
	if err != nil {
		t.Fatalf("Error while closing websocket: %v", err)
	}
	mockedClock.Add(time.Duration(2) * time.Second) // Has to be 2 seconds, to trigger two messages; first message always gets through for some reason

	// Assert
	t.Run("NumberOfPointsReceived", func(t *testing.T) {
		got := string(data)
		want := `NumberOfPoints="2"`
		if got != want {
			t.Errorf("Incorrect message, got: %v, want: %v", got, want)
		}
	})

	t.Run("UnSubscribedSuccesfully", func(t *testing.T) {
		if mockedHub.Subscribed != 0 {
			t.Error("Did not subscribe succesfully")
		}
	})

}
