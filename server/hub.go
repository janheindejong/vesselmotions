package main

import (
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
