package main

import (
	"log"
)

// Hub is the implementation of the PubSub pattern, responsible for
// receiving datapoints from various sources, and forwarding them to
// subscribers in the form of channels. It will also be responsible for
// error handling, such as full queues.
type Hub[T any] struct {
	subscribers map[chan T]string
}

func (hub *Hub[T]) Subscribe(bufferSize int) chan T {
	c := make(chan T, bufferSize)
	hub.subscribers[c] = ""
	log.Print("Added subscriber, number of subscriptions is ", len(hub.subscribers))
	return c
}

func (hub *Hub[T]) UnSubscribe(c chan T) {
	delete(hub.subscribers, c)
	log.Print("Removed subscriber, number of subscriptions is ", len(hub.subscribers))
}

func (hub *Hub[T]) Broadcast(msg T) {
	for subscriber := range hub.subscribers {
		select {
		case subscriber <- msg:
			continue
		default:
			log.Print(`Buffer full, message dropped`)
		}

	}
}
