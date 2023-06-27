package main

import (
	"fmt"
	"github.com/smartwalle/sse"
	"log"
	"net/http"
	"time"
)

// https://devtest.run/sse.html#

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var sServer = sse.New()

	sServer.OnStreamOpen(func(stream *sse.Stream) {
		log.Println("open...")

		for {
			if err := stream.Send(&sse.Event{Id: "111", Data: "hahaha"}); err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(time.Second)
		}
	})

	sServer.OnStreamClose(func(stream *sse.Stream) {
		log.Println("close...")
	})

	http.HandleFunc("/sse", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Sec-Websocket-Key, Connection, Sec-Websocket-Version, Sec-Websocket-Extensions, Upgrade, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		sServer.ServeHTTP(writer, request)
	})

	http.ListenAndServe(":9911", nil)
}
