package sse

import (
	"bytes"
	"strconv"
	"strings"
)

var replacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")

type Event struct {
	ID    string
	Event string
	Retry uint
	Data  string
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
		buf.WriteString(strconv.FormatUint(uint64(e.Retry), 10))
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
