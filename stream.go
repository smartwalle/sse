package sse

import (
	"errors"
	"net/http"
	"sync"
)

var ErrClosed = errors.New("stream closed")
var ErrUnsupported = errors.New("server-send events unsupported")

type Stream struct {
	writer    http.ResponseWriter
	flusher   http.Flusher
	request   *http.Request
	closed    chan struct{}
	closeOnce sync.Once
}

func Upgrade(writer http.ResponseWriter, request *http.Request) (*Stream, error) {
	var flusher, ok = writer.(http.Flusher)
	if !ok {
		return nil, ErrUnsupported
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	var stream = &Stream{}
	stream.writer = writer
	stream.flusher = flusher
	stream.request = request
	stream.closed = make(chan struct{})
	return stream, nil
}

func (s *Stream) Wait() {
	for {
		select {
		case <-s.request.Context().Done():
			s.Close()
			return
		case <-s.closed:
			return
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

func (s *Stream) Write(data []byte) (int, error) {
	n, err := s.writer.Write(data)
	if err != nil {
		return n, err
	}
	select {
	case <-s.request.Context().Done():
		return 0, ErrClosed
	default:
		s.flusher.Flush()
	}
	return n, nil
}

func (s *Stream) Send(event Event) (err error) {
	var data = Encode(event)
	if _, err = s.Write(data); err != nil {
		return err
	}
	return nil
}
