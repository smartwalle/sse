package sse

import (
	"errors"
	"fmt"
	"io"
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

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	var stream = newStream(request)

	defer func() {
		if this.onClose != nil {
			this.onClose(stream)
		}
	}()

	go func() {
		this.onOpen(stream)
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
			this.encode(writer, event)
			flusher.Flush()
		}
	}
}

func (this *Server) encode(w io.Writer, event *Event) {
	if len(event.Id) > 0 {
		fmt.Fprintf(w, "id: %s\n", event.Id)
	}

	if len(event.Event) > 0 {
		fmt.Fprintf(w, "event: %s\n", event.Event)
	}

	if len(event.Data) > 0 {
		fmt.Fprintf(w, "data: %s\n", event.Data)
	}

	if event.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", event.Retry)
	}

	fmt.Fprintf(w, "\n")
}
