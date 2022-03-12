package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

type MockSubscriber struct {
	channel                chan DataPoint
	numberOfPointsReceived int
	latestValue            float64
}

func (sub *MockSubscriber) Process() {
	close(sub.channel)
	for point := range sub.channel {
		sub.latestValue = point.Value
		sub.numberOfPointsReceived++
	}
}

func NewMockSubscriber(size int) *MockSubscriber {
	return &MockSubscriber{
		channel:                make(chan DataPoint, size),
		numberOfPointsReceived: 0,
	}
}

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestHub(t *testing.T) {
	// Arrange
	hub := Hub{
		subscribers: map[chan DataPoint]string{},
	}

	subscribers := []*MockSubscriber{
		NewMockSubscriber(1),
		NewMockSubscriber(2),
	}

	for _, sub := range subscribers {
		hub.Subscribe(sub.channel)
	}

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
	for _, datapoint := range datapoints {
		hub.Broadcast(datapoint)
	}

	for _, sub := range subscribers {
		hub.UnSubscribe(sub.channel)
		sub.Process()
	}

	// Assert
	testCasesNumberOfPointsReceived := map[string]struct {
		sub  *MockSubscriber
		want int
	}{
		"Receiver0Length": {sub: subscribers[0], want: 1},
		"Receiver1Length": {sub: subscribers[1], want: 2},
	}

	for name, tc := range testCasesNumberOfPointsReceived {
		t.Run(name, func(t *testing.T) {
			got := tc.sub.numberOfPointsReceived
			if got != tc.want {
				t.Errorf("Incorrect channel length, got: %d, want: %d", got, tc.want)
			}
		})
	}

	testCasesLatestValueReceived := map[string]struct {
		sub  *MockSubscriber
		want float64
	}{
		"Receiver0LatestValue": {sub: subscribers[0], want: datapoints[0].Value},
		"Receiver1LatestValue": {sub: subscribers[1], want: datapoints[1].Value},
	}

	for name, tc := range testCasesLatestValueReceived {
		t.Run(name, func(t *testing.T) {
			got := tc.sub.latestValue
			if got != tc.want {
				t.Errorf("Incorrect latest value, got: %f, want: %f", got, tc.want)
			}
		})
	}

	t.Run("AllUnSubscribed", func(t *testing.T) {
		got := len(hub.subscribers)
		if got != 0 {
			t.Errorf("Too many subscribers left, got: %d, want: 0", got)

		}
	})
}
