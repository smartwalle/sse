package sse

import (
	"errors"
	"net/http"
	"sync"
)

var ErrClosed = errors.New("stream closed")

type Stream struct {
	writer    http.ResponseWriter
	flusher   http.Flusher
	request   *http.Request
	encoder   func(Event) []byte
	closed    chan struct{}
	closeOnce sync.Once
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
	select {
	case <-s.request.Context().Done():
		return 0, ErrClosed
	default:
		n, err := s.writer.Write(data)
		if err != nil {
			return n, err
		}
		s.flusher.Flush()
		return n, nil
	}
}

func (s *Stream) Send(event Event) (err error) {
	var data = s.encoder(event)
	if _, err = s.Write(data); err != nil {
		return err
	}
	return nil
}
