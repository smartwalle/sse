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

func (this *Server) Serve(id, tag string, writer http.ResponseWriter, request *http.Request) error {
	var flusher, ok = writer.(http.Flusher)
	if !ok {
		return ErrUnsupported
	}

	writer.Header().Set("Content-Type", "text/events-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	this.mu.Lock()
	var stream = this.streams[id]
	if stream == nil {
		stream = newStream(id)
		this.streams[stream.id] = stream
	}
	this.mu.Unlock()

	var subscriber = stream.addSubscriber(tag)

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	defer func() {
		<-request.Context().Done()
		stream.removeSubscriber(subscriber)
	}()

	for event := range subscriber.events {
		if event == nil {
			return nil
		}
		writer.Write(Encode(event))
		flusher.Flush()
	}

	return nil
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

func (this *Server) StreamExists(id string) bool {
	this.mu.Lock()
	var _, ok = this.streams[id]
	this.mu.Unlock()
	return ok
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

func (this *Server) Send(id string, event *Event) error {
	this.mu.Lock()
	var stream = this.streams[id]
	this.mu.Unlock()

	if stream == nil {
		return ErrNotFound
	}

	select {
	case <-stream.closed:
		return ErrClosed
	case stream.events <- event:
	}
	return nil
}
