package sse

import (
	"errors"
	"sync"
)

var ErrClosed = errors.New("stream already closed")
var ErrNotFound = errors.New("stream not found")

type Stream struct {
	id          string
	closed      chan struct{}
	closeOnce   sync.Once
	events      chan *Event
	subscribers map[string]*Subscriber
}

func newStream(id string) *Stream {
	var nStream = &Stream{}
	nStream.id = id
	nStream.closed = make(chan struct{})
	nStream.events = make(chan *Event)
	nStream.subscribers = make(map[string]*Subscriber)
	go nStream.run()
	return nStream
}

func (this *Stream) Id() string {
	return this.id
}

func (this *Stream) run() {
	for {
		select {
		case event := <-this.events:
			for _, sub := range this.subscribers {
				sub.events <- event
			}
		case <-this.closed:
			this.removeAllSubscriber()
			return
		}
	}
}

func (this *Stream) close() {
	this.closeOnce.Do(func() {
		close(this.closed)
	})
}

func (this *Stream) addSubscriber(tag string) *Subscriber {
	var subscriber = this.subscribers[tag]
	if subscriber != nil {
		delete(this.subscribers, tag)
		subscriber.clean()
	}

	subscriber = newSubscriber(tag)
	this.subscribers[tag] = subscriber

	return subscriber
}

func (this *Stream) removeSubscriber(subscriber *Subscriber) {
	for _, sub := range this.subscribers {
		if sub == subscriber {
			delete(this.subscribers, sub.tag)
			subscriber.clean()
		}
	}
}

func (this *Stream) removeAllSubscriber() {
	for tag := range this.subscribers {
		this.subscribers[tag].clean()
		delete(this.subscribers, tag)
	}
}

func (this *Stream) removable() bool {
	return len(this.subscribers) == 0
}
