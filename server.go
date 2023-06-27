package sse

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrUnsupported = errors.New("server-Send events unsupported")

type Handler func(stream *Stream)

type Server struct {
	onOpen  Handler
	onClose Handler
}

func New() *Server {
	var nServer = &Server{}
	nServer.onOpen = func(stream *Stream) {
		panic(`open stream handler should be specified.
var s = sse.Server()
s.OnStreamOpen(func(stream *sse.Stream) {
})`)
	}
	return nServer
}

func (this *Server) OnStreamOpen(handler Handler) {
	this.onOpen = handler
}

func (this *Server) OnStreamClose(handler Handler) {
	this.onClose = handler
}

func (this *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var flusher, ok = writer.(http.Flusher)
	if !ok {
		fmt.Fprintln(writer, ErrUnsupported.Error())
		return
	}

	writer.Header().Set("Content-Type", "text/events-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	var stream = newStream(request)

	var ready = make(chan struct{})

	defer func() {
		if this.onClose != nil {
			<-ready
			this.onClose(stream)
		}
	}()

	go func() {
		this.onOpen(stream)
		close(ready)
	}()

	for {
		select {
		case <-request.Context().Done():
			stream.Close()
			return
		case <-stream.closed:
			return
		case event := <-stream.events:
			if event == nil {
				return
			}
			writer.Write(Encode(event))
			flusher.Flush()
		}
	}
}
