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

func (s *Stream) Request() *http.Request {
	return s.request
}

func (s *Stream) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
	})
	return nil
}

func (s *Stream) Send(event *Event) error {
	select {
	case <-s.closed:
		return ErrClosed
	default:
		select {
		case <-s.closed:
			return ErrClosed
		case s.events <- event:
			return nil
		}
	}
}
