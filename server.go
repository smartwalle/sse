package sse

import (
	"errors"
	"net/http"
	"sync"
)

var ErrUnsupported = errors.New("server-Send events unsupported")

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

	var nStream = newStream(streamId, writer, flusher)
	this.addStream(nStream)

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	defer func() {
		this.RemoveStream(streamId)
	}()

	for {
		select {
		case <-request.Context().Done():
			return request.Context().Err()
		case <-nStream.closed:
			return nil
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

func (this *Server) RemoveStream(id string) {
	this.mu.Lock()
	var stream = this.streams[id]
	delete(this.streams, id)
	this.mu.Unlock()

	if stream != nil {
		stream.close()
	}
}

func (this *Server) Send(id string, event []byte) error {
	this.mu.Lock()
	var stream = this.streams[id]
	this.mu.Unlock()

	if stream == nil {
		return ErrNotFound
	}

	return stream.Write(event)
}
