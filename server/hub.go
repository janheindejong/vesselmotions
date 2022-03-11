package main

import (
	"errors"
	"log"
	"time"
)

type DataPoint struct {
	Value     float64   `json:"Value"`
	Timestamp time.Time `json:"Timestamp"`
	Id        string    `json:"Id"`
}

type Hub struct {
	subscribers map[chan DataPoint]string
}

func (hub *Hub) Subscribe(sub chan DataPoint) {
	hub.subscribers[sub] = ""
	log.Print("Added subscriber, number of subscriptions is ", len(hub.subscribers))
}

func (hub *Hub) UnSubscribe(sub chan DataPoint) {
	delete(hub.subscribers, sub)
	log.Print("Removed subscriber, number of subscriptions is ", len(hub.subscribers))
}

func (hub *Hub) Broadcast(point DataPoint) {
	for subscriber := range hub.subscribers {
		select {
		case subscriber <- point:
			// Message sent
		default:
			log.Print(`Buffer full, message dropped`)
		}

	}
}

func emptyBuffer(channel chan DataPoint) *[]DataPoint {
	datapoints := make([]DataPoint, 0)
	for {
		select {
		case NewDataPoint := <-channel:
			datapoints = append(datapoints, NewDataPoint)
		default:
			return &datapoints
		}
	}
}

type Subscriber struct {
	buffer chan DataPoint
}

func (sub *Subscriber) Notify(point DataPoint) error {
	select {
	case sub.buffer <- point:
		return nil
	default:
		return errors.New("buffer full")
	}
}

func (sub *Subscriber) EmptyBuffer() *[]DataPoint {
	datapoints := make([]DataPoint, 0)
	for {
		select {
		case NewDataPoint := <-sub.buffer:
			datapoints = append(datapoints, NewDataPoint)
		default:
			return &datapoints
		}
	}
}
