package main

import (
	"github.com/smartwalle/sse"
	"log"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var sServer = sse.New()

	http.HandleFunc("/sse", func(writer http.ResponseWriter, request *http.Request) {
		sServer.Serve("test", writer, request)
	})

	go func() {
		var idx = 0
		for {
			var stream = sServer.GetStream("test")
			if stream != nil {
				stream.Send([]byte("hahahah"))
			}
			idx++
			time.Sleep(time.Second)
			if idx > 10 {
				sServer.CloseStream("test")
				return
			}
		}
	}()

	http.ListenAndServe(":9911", nil)
}
