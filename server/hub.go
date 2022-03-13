package main

import (
	"log"
)

// Hub is the implementation of the PubSub pattern, responsible for
// receiving datapoints from various sources, and forwarding them to
// subscribers in the form of channels. It will also be responsible for
// error handling, such as full queues.
type Hub struct {
	subscribers map[chan DataPoint]string
}

func (hub *Hub) Subscribe(bufferSize int) chan DataPoint {
	c := make(chan DataPoint, bufferSize)
	hub.subscribers[c] = ""
	log.Print("Added subscriber, number of subscriptions is ", len(hub.subscribers))
	return c
}

func (hub *Hub) UnSubscribe(c chan DataPoint) {
	delete(hub.subscribers, c)
	log.Print("Removed subscriber, number of subscriptions is ", len(hub.subscribers))
}

func (hub *Hub) Broadcast(point DataPoint) {
	for subscriber := range hub.subscribers {
		select {
		case subscriber <- point:
			continue
		default:
			log.Print(`Buffer full, message dropped`)
		}

	}
}
