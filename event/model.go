package event

type NewHolderEvent struct {
	CurrentHolderUserId string
	PrevHolderUserId    string
	AuthorUserId        string
}
