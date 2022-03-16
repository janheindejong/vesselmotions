package main

import (
	"math"
	"time"
)

// MockSensorDataPublisher is responsible for generating random sensor data
type MockSensorDataPublisher struct {
	hub *Hub[DataPoint]
}

type DataPoint struct {
	Value     float64   `json:"Value"`
	Timestamp time.Time `json:"Timestamp"`
	Id        string    `json:"Id"`
}

func (publisher *MockSensorDataPublisher) RunForever() {
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
