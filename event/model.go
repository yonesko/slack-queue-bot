package event

import "time"

type NewHolderEvent struct {
	CurrentHolderUserId string
	PrevHolderUserId    string
	AuthorUserId        string
	ts                  time.Time
}
