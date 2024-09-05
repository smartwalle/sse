package sse

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrUnsupported = errors.New("server-Send events unsupported")

type Handler func(stream *Stream)

type Server struct {
	onOpen   Handler
	onClose  Handler
	replacer *strings.Replacer
}

func New() *Server {
	var nServer = &Server{}
	nServer.onOpen = func(stream *Stream) {
		panic(`open stream handler should be specified.
var s = sse.Server()
s.OnStreamOpen(func(stream *sse.Stream) {
})`)
	}
	nServer.replacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")
	return nServer
}

func (s *Server) OnStreamOpen(handler Handler) {
	s.onOpen = handler
}

func (s *Server) OnStreamClose(handler Handler) {
	s.onClose = handler
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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
		if s.onClose != nil {
			s.onClose(stream)
		}
	}()

	go func() {
		s.onOpen(stream)
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
			s.encode(writer, event)
			flusher.Flush()
		}
	}
}

func (s *Server) encode(w io.Writer, event *Event) {
	if len(event.Id) > 0 {
		fmt.Fprintf(w, "id: %s\n", s.replacer.Replace(event.Id))
	}
	if len(event.Event) > 0 {
		fmt.Fprintf(w, "event: %s\n", s.replacer.Replace(event.Event))
	}
	if event.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", event.Retry)
	}
	if len(event.Data) > 0 {
		fmt.Fprintf(w, "data: %s\n\n", s.replacer.Replace(event.Data))
	}
}
