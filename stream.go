package sse

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var ErrClosed = errors.New("stream already closed")
var ErrNotFound = errors.New("stream not found")

type Stream struct {
	id        string
	closed    chan struct{}
	closeOnce sync.Once

	writer  http.ResponseWriter
	flusher http.Flusher
}

func newStream(id string, writer http.ResponseWriter, flusher http.Flusher) *Stream {
	var nStream = &Stream{}
	nStream.id = id
	nStream.closed = make(chan struct{})
	nStream.writer = writer
	nStream.flusher = flusher
	return nStream
}

func (this *Stream) Id() string {
	return this.id
}

func (this *Stream) close() {
	this.closeOnce.Do(func() {
		close(this.closed)
	})
}

func (this *Stream) Write(event []byte) error {
	_, err := fmt.Fprintln(this.writer, event)
	if err != nil {
		return err
	}
	this.flusher.Flush()
	return nil
}
