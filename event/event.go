package event

type QueueChangedEvent interface {
}

type QueueChangedEventBus interface {
	Send(queueChangedEvent QueueChangedEvent)
}
