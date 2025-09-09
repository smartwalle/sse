package sse

import (
	"context"
	"errors"
	"net/http"
)

var ErrUnsupported = errors.New("streaming unsupported")

func Upgrade(ctx context.Context, writer http.ResponseWriter, request *http.Request) (*Stream, error) {
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
	stream.ctx = ctx
	stream.writer = writer
	stream.flusher = flusher
	stream.request = request
	stream.encoder = Encode
	stream.closed = make(chan struct{})
	return stream, nil
}
