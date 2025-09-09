package sse

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

var ErrClosed = errors.New("stream closed")

type Stream struct {
	ctx       context.Context
	writer    http.ResponseWriter
	flusher   http.Flusher
	request   *http.Request
	encoder   func(Event) []byte
	closed    chan struct{}
	closeOnce sync.Once
}

func (s *Stream) Wait() {
	select {
	case <-s.ctx.Done():
		s.Close()
		return
	case <-s.closed:
		return
	}
}

func (s *Stream) Request() *http.Request {
	return s.request
}

func (s *Stream) Closed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

func (s *Stream) Close() error {
	s.closeOnce.Do(func() {
		if s.closed != nil {
			close(s.closed)
		}
	})
	return nil
}

func (s *Stream) Write(data []byte) (int, error) {
	select {
	case <-s.closed:
		return 0, ErrClosed
	case <-s.ctx.Done():
		return 0, ErrClosed
	default:
		if s.writer == nil {
			return 0, ErrClosed
		}
		n, err := s.writer.Write(data)
		if err != nil {
			return n, err
		}
		if s.flusher != nil {
			s.flusher.Flush()
		}
		return n, nil
	}
}

func (s *Stream) Send(event Event) (err error) {
	if s.encoder == nil {
		return ErrClosed
	}

	var data = s.encoder(event)
	if _, err = s.Write(data); err != nil {
		return err
	}
	return nil
}
