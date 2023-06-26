package sse

import (
	"errors"
	"sync"
)

var ErrClosed = errors.New("stream already closed")

type Stream struct {
	id        string
	closed    chan struct{}
	closeOnce sync.Once
	events    chan []byte
}

func newStream(id string) *Stream {
	var nStream = &Stream{}
	nStream.id = id
	nStream.closed = make(chan struct{})
	nStream.events = make(chan []byte, 10)
	return nStream
}

func (this *Stream) Id() string {
	return this.id
}

func (this *Stream) close() {
	this.closeOnce.Do(func() {
		close(this.closed)
		close(this.events)
	})
}

func (this *Stream) Send(data []byte) error {
	select {
	case <-this.closed:
		return ErrClosed
	default:
		select {
		case <-this.closed:
			return ErrClosed
		case this.events <- data:
			return nil
		}
	}
	return nil
}
