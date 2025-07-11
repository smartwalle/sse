package main

import (
	"github.com/smartwalle/sse"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:9091/sse", nil)
	if err != nil {
		log.Println("创建 Request 异常:", err)
		return
	}

	var client = sse.NewClient(req)
	client.OnEvent(func(event *sse.Event) {
		log.Println("接收到 Event:", event)
	})

	log.Println(client.Connect())
}
