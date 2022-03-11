package main

import (
	"math"
	"time"
)

type MockPublisher struct {
	hub *Hub
}

func (publisher *MockPublisher) Run() {
	for {
		point := createMockDataPoint()
		publisher.hub.Broadcast(*point)
		time.Sleep(time.Millisecond * time.Duration(100))
	}
}

func createMockDataPoint() *DataPoint {
	TimeSeconds := float64(time.Now().UnixMilli()) / 1e3
	Value := math.Sin(2 * math.Pi * TimeSeconds / 10) // Mock implementation
	point := DataPoint{
		Value:     Value,
		Timestamp: time.Now(),
	}
	return &point
}
