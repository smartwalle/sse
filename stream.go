package sse

import (
	"errors"
	"net/http"
	"sync"
)

var ErrClosed = errors.New("stream already closed")

type Stream struct {
	request   *http.Request
	events    chan *Event
	closed    chan struct{}
	closeOnce sync.Once
}

func newStream(req *http.Request) *Stream {
	var nStream = &Stream{}
	nStream.request = req
	nStream.events = make(chan *Event, 8)
	nStream.closed = make(chan struct{})
	return nStream
}

func (this *Stream) Request() *http.Request {
	return this.request
}

func (this *Stream) Close() error {
	this.closeOnce.Do(func() {
		close(this.closed)
	})
	return nil
}

func (this *Stream) Send(event *Event) error {
	select {
	case <-this.closed:
		return ErrClosed
	default:
		select {
		case <-this.closed:
			return ErrClosed
		case this.events <- event:
			return nil
		}
	}
}
