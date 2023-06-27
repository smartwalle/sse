package sse

type Event struct {
	Id    string
	Event string
	Retry uint
	Data  string
}
