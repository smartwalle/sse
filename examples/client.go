package main

import (
	"context"
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
	client.OnCheck(func(ctx context.Context, response *http.Response) error {
		return nil
	})
	client.OnEvent(func(ctx context.Context, event *sse.Event) error {
		log.Printf("接收到 Event: %+v\n", event)
		return nil
	})

	log.Println(client.Connect(context.Background()))
}
