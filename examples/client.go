package main

import (
	"log"
	"net/http"

	"github.com/smartwalle/sse"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:9091/sse", nil)
	if err != nil {
		log.Println("创建 Request 异常:", err)
		return
	}

	var client = sse.NewClient(req)
	client.OnEvent(func(event *sse.Event) error {
		log.Println("接收到 Event:", event)
		return nil
	})

	log.Println(client.Connect())
}
