package sse

type Event struct {
	Data []byte
}

func Encode(event *Event) []byte {
	return event.Data
}
