package sse

import (
	"errors"
	"github.com/smartwalle/nsync"
)

var ErrClosed = errors.New("stream already closed")
var ErrNotFound = errors.New("stream not found")

type Stream struct {
	id          string
	quit        *nsync.Event
	events      chan *Event
	subscribers map[string]*Subscriber

	register   chan *Subscriber
	deregister chan *Subscriber
}

func newStream(id string) *Stream {
	var nStream = &Stream{}
	nStream.id = id
	nStream.quit = nsync.NewEvent()
	nStream.events = make(chan *Event)
	nStream.subscribers = make(map[string]*Subscriber)
	nStream.register = make(chan *Subscriber)
	nStream.deregister = make(chan *Subscriber)
	go nStream.run()
	return nStream
}

func (this *Stream) Id() string {
	return this.id
}

func (this *Stream) run() {
	for {
		select {
		case subscriber := <-this.register:
			if ele := this.subscribers[subscriber.tag]; ele != nil {
				delete(this.subscribers, ele.tag)
				ele.close()
			}
			this.subscribers[subscriber.tag] = subscriber
		case subscriber := <-this.deregister:
			if ele := this.subscribers[subscriber.tag]; ele == subscriber {
				delete(this.subscribers, subscriber.tag)
			}
			subscriber.close()
		case event := <-this.events:
			for _, ele := range this.subscribers {
				if event.Tag != "" && event.Tag != ele.tag {
					continue
				}

				select {
				case <-ele.quit.Done():
					continue
				case ele.events <- event:
				}
			}
		case <-this.quit.Done():
			for _, ele := range this.subscribers {
				ele.close()
			}
			this.subscribers = nil
			return
		}
	}
}

func (this *Stream) close() {
	if this.quit.HasFired() {
		return
	}
	this.quit.Fire()
	close(this.events)
	close(this.register)
	close(this.deregister)
}

func (this *Stream) addSubscriber(tag string) *Subscriber {
	var subscriber = newSubscriber(tag)
	select {
	case <-this.quit.Done():
		return nil
	case this.register <- subscriber:
	}
	return subscriber
}

func (this *Stream) removeSubscriber(subscriber *Subscriber) {
	if subscriber != nil {
		select {
		case <-this.quit.Done():
		case <-subscriber.quit.Done():
		case this.deregister <- subscriber:
			select {
			case <-this.quit.Done():
			case <-subscriber.quit.Done():
			}
		}
	}
}

func (this *Stream) removable() bool {
	select {
	case <-this.quit.Done():
		return true
	default:
		return len(this.subscribers) == 0
	}
}
