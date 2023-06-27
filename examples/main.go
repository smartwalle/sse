package main

import (
	"fmt"
	"github.com/smartwalle/sse"
	"log"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var sServer = sse.New()

	http.HandleFunc("/sse", func(writer http.ResponseWriter, request *http.Request) {

		fmt.Println(request.URL.String())

		request.ParseForm()

		var id = request.Form.Get("id")
		var tag = request.Form.Get("tag")

		sServer.Serve(id, tag, writer, request)
	})

	go func() {
		var idx = 0
		for {

			sServer.Send("test", &sse.Event{Data: []byte("hello")})

			idx++
			time.Sleep(time.Second)
		}
	}()

	http.ListenAndServe(":9911", nil)
}
