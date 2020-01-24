package event

type NewHolderEvent struct {
	CurrentHolderUserId string
}

type QueueChangedEventBus interface {
	Send(event interface{})
}
