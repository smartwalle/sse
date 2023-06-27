package sse

type Event struct {
	Tag  string
	Data []byte
}

func Encode(event *Event) []byte {
	return event.Data
}
