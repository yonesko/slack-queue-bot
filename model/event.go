package model

import "time"

type NewHolderEvent struct {
	CurrentHolderUserId string
	PrevHolderUserId    string
	AuthorUserId        string
	SecondUserId        string
	Ts                  time.Time
}
