package sse

import "github.com/smartwalle/nsync"

type Subscriber struct {
	tag    string
	quit   *nsync.Event
	events chan *Event
}

func newSubscriber(tag string) *Subscriber {
	var nSubscriber = &Subscriber{}
	nSubscriber.tag = tag
	nSubscriber.quit = nsync.NewEvent()
	nSubscriber.events = make(chan *Event)
	return nSubscriber
}

func (this *Subscriber) close() {
	if this.quit.HasFired() {
		return
	}
	this.quit.Fire()
	close(this.events)
}
