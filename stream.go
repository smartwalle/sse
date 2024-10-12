package sse

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var ErrClosed = errors.New("stream already closed")
var ErrUnsupported = errors.New("server-send events unsupported")

var replacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")

type Stream struct {
	writer    http.ResponseWriter
	flusher   http.Flusher
	request   *http.Request
	events    chan *Event
	closed    chan struct{}
	closeOnce sync.Once
}

func NewStream(writer http.ResponseWriter, request *http.Request) (*Stream, error) {
	var flusher, ok = writer.(http.Flusher)
	if !ok {
		return nil, ErrUnsupported
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	var nStream = &Stream{}
	nStream.writer = writer
	nStream.flusher = flusher
	nStream.request = request
	nStream.events = make(chan *Event, 8)
	nStream.closed = make(chan struct{})
	return nStream, nil
}

func (s *Stream) Run() {
	for {
		select {
		case <-s.request.Context().Done():
			s.Close()
			return
		case <-s.closed:
			return
		case event := <-s.events:
			if event != nil {
				write(s.writer, s.flusher, event)
			}
		}
	}
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
	if event == nil {
		return nil
	}

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

func write(writer io.Writer, flusher http.Flusher, event *Event) {
	if len(event.Id) > 0 {
		fmt.Fprintf(writer, "id: %s\n", replacer.Replace(event.Id))
	}
	if len(event.Event) > 0 {
		fmt.Fprintf(writer, "event: %s\n", replacer.Replace(event.Event))
	}
	if event.Retry > 0 {
		fmt.Fprintf(writer, "retry: %d\n", event.Retry)
	}
	if len(event.Data) > 0 {
		fmt.Fprintf(writer, "data: %s\n\n", replacer.Replace(event.Data))
	}
	flusher.Flush()
}
