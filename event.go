package sse

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var replacer = strings.NewReplacer(
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
)

type Event struct {
	ID    string
	Event string
	Retry int
	Data  string
}

func (e Event) String() string {
	return fmt.Sprintf("Event{ID: %s, Event: %s, Retry: %d, Data: %s}", e.ID, e.Event, e.Retry, e.Data)
}

func Encode(event Event) []byte {
	var buffer = bytes.NewBuffer([]byte{})
	EncodeToBuffer(event, buffer)
	return buffer.Bytes()
}

func EncodeToBuffer(event Event, buffer *bytes.Buffer) {
	var size int
	if event.ID != "" {
		event.ID = replacer.Replace(event.ID)
		size += 4 + len(event.ID) + 1 // "id: " + ID + "\n"
	}
	if event.Event != "" {
		event.Event = replacer.Replace(event.Event)
		size += 7 + len(event.Event) + 1 // "event: " + Event + "\n"
	}
	if event.Retry > 0 {
		size += 7 + 19 + 1 // "retry: " + math.MaxInt + "\n"
	}
	if event.Data != "" {
		event.Data = replacer.Replace(event.Data)
		size += 6 + len(event.Data) + 1 // "data: " + Data + "\n"
	}
	size += 1 // 最后的空行 "\n"

	buffer.Grow(size)

	if event.ID != "" {
		buffer.WriteString("id: ")
		buffer.WriteString(event.ID)
		buffer.WriteByte('\n')
	}

	if event.Event != "" {
		buffer.WriteString("event: ")
		buffer.WriteString(event.Event)
		buffer.WriteByte('\n')
	}

	if event.Retry > 0 {
		buffer.WriteString("retry: ")
		buffer.WriteString(strconv.Itoa(event.Retry))
		buffer.WriteByte('\n')
	}

	if event.Data != "" {
		buffer.WriteString("data: ")
		buffer.WriteString(event.Data)
		buffer.WriteByte('\n')
	}
	buffer.WriteByte('\n')
}
