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

			var event = &sse.Event{}
			if idx%2 == 0 {
				event.Tag = "1"
			} else {
				event.Tag = "2"
			}

			if idx%3 == 0 {
				event.Tag = ""
			}

			event.Data = []byte(fmt.Sprintf("hello %d ", idx))

			sServer.Send("test", event)

			idx++
			time.Sleep(time.Second)
		}
	}()

	http.ListenAndServe(":9911", nil)
}
