package sse

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var ErrUnsupported = errors.New("sse unsupported")

type Server struct {
	mu      sync.Mutex
	streams map[string]*Stream
}

func New() *Server {
	var nServer = &Server{}
	nServer.streams = make(map[string]*Stream)
	return nServer
}

func (this *Server) Serve(streamId string, writer http.ResponseWriter, request *http.Request) error {
	var flusher, ok = writer.(http.Flusher)
	if !ok {
		return ErrUnsupported
	}

	writer.Header().Set("Content-Type", "text/events-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	var nStream = newStream(streamId)
	this.addStream(nStream)

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	defer func() {
		this.CloseStream(streamId)
	}()

	for {
		select {
		case <-request.Context().Done():
			return request.Context().Err()
		case <-nStream.closed:
			return nil
		case event, ok := <-nStream.events:
			if !ok {
				return nil
			}

			fmt.Fprintln(writer, event)
			flusher.Flush()
		}
	}
}

func (this *Server) Close() error {
	this.mu.Lock()
	for id := range this.streams {
		this.streams[id].close()
		delete(this.streams, id)
	}
	this.mu.Unlock()
	return nil
}

func (this *Server) addStream(stream *Stream) {
	this.mu.Lock()
	this.streams[stream.id] = stream
	this.mu.Unlock()
}

func (this *Server) GetStream(id string) *Stream {
	this.mu.Lock()
	var stream = this.streams[id]
	this.mu.Unlock()
	return stream
}

func (this *Server) CloseStream(id string) {
	this.mu.Lock()
	var stream = this.streams[id]
	delete(this.streams, id)
	this.mu.Unlock()

	if stream != nil {
		stream.close()
	}
}
