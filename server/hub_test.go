package main

import (
	"testing"
)

type MockSubscriber struct {
	channel                chan int
	bufferSize             int
	numberOfPointsReceived int
	latestValue            int
}

func (sub *MockSubscriber) Process() {
	close(sub.channel)
	for point := range sub.channel {
		sub.latestValue = point
		sub.numberOfPointsReceived++
	}
}

func NewMockSubscriber(bufferSize int) *MockSubscriber {
	return &MockSubscriber{
		numberOfPointsReceived: 0,
		bufferSize:             bufferSize,
	}
}

func TestHub(t *testing.T) {

	// Arrange
	hub := Hub[int]{
		subscribers: map[chan int]string{},
	}

	subscribers := []*MockSubscriber{
		NewMockSubscriber(1),
		NewMockSubscriber(2),
	}

	for _, sub := range subscribers {
		sub.channel = hub.Subscribe(sub.bufferSize)
	}

	messages := []int{0, 1}

	// Act
	for _, msg := range messages {
		hub.Broadcast(msg)
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
		want int
	}{
		"Receiver0LatestValue": {sub: subscribers[0], want: messages[0]},
		"Receiver1LatestValue": {sub: subscribers[1], want: messages[1]},
	}

	for name, tc := range testCasesLatestValueReceived {
		t.Run(name, func(t *testing.T) {
			got := tc.sub.latestValue
			if got != tc.want {
				t.Errorf("Incorrect latest value, got: %d, want: %d", got, tc.want)
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
