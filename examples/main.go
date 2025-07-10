package main

import (
	"fmt"
	"github.com/smartwalle/sse"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	http.HandleFunc("/sse", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Sec-Websocket-Key, Connection, Sec-Websocket-Version, Sec-Websocket-Extensions, Upgrade, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		var stream, err = sse.Upgrade(writer, request)
		if err != nil {
			log.Println("建立 Stream 异常:", err)
			return
		}

		log.Println("建立 Stream 成功")

		go func() {
			var idx = 1
			for {
				if err := stream.Send(sse.Event{ID: fmt.Sprintf("%d", idx), Data: strings.Repeat("hello,", 1000), Event: time.Now().Format(time.RFC3339)}); err != nil {
					log.Println("推送数据异常：", err)
					return
				}

				idx++
				//if idx == 10 {
				//	stream.Close()
				//}
				time.Sleep(time.Second * 1)
			}
		}()
		stream.Wait()
		log.Println("关闭 Stream")
	})

	http.ListenAndServe(":9091", nil)
}
