package main

import (
	"github.com/smartwalle/sse"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var sServer = sse.New()

	sServer.OnStreamOpen(func(stream *sse.Stream) {
		log.Println("open...")
	})

	sServer.OnStreamClose(func(stream *sse.Stream) {
		log.Println("close...")
	})

	http.HandleFunc("/sse", func(writer http.ResponseWriter, request *http.Request) {
		sServer.ServeHTTP(writer, request)
	})

	http.ListenAndServe(":9911", nil)
}
