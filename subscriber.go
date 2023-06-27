package sse

type Subscriber struct {
	tag    string
	events chan *Event
}

func newSubscriber(tag string) *Subscriber {
	var nSubscriber = &Subscriber{}
	nSubscriber.tag = tag
	nSubscriber.events = make(chan *Event)
	return nSubscriber
}

func (this *Subscriber) clean() {
	close(this.events)
}
