package model

import "time"

type NewHolderEvent struct {
	CurrentHolderUserId string
	PrevHolderUserId    string
	AuthorUserId        string
	Ts                  time.Time
}

type NewSecondEvent struct {
	CurrentSecondUserId string
}

type DeletedEvent struct {
	AuthorUserId string
}
