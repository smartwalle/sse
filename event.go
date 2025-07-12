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
	"\\", "\\\\",
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

func Encode(e Event) []byte {
	var buf bytes.Buffer

	if e.ID != "" {
		buf.WriteString("id: ")
		buf.WriteString(replacer.Replace(e.ID))
		buf.WriteByte('\n')
	}

	if e.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(replacer.Replace(e.Event))
		buf.WriteByte('\n')
	}

	if e.Retry > 0 {
		buf.WriteString("retry: ")
		buf.WriteString(strconv.Itoa(e.Retry))
		buf.WriteByte('\n')
	}

	if e.Data != "" {
		buf.WriteString("data: ")
		buf.WriteString(replacer.Replace(e.Data))
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}
