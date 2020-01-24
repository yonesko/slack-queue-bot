package mock

type QueueChangedEventBus struct {
	Inbox []interface{}
}

func (q *QueueChangedEventBus) Send(event interface{}) {
	q.Inbox = append(q.Inbox, event)
}
