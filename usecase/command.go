package usecase

type Command struct {
	AuthorUserId string
	Data         interface{}
}

type AddCommand struct {
	ToAddUserId string
}

type DelCommand struct {
	ToDelUserId string
}

type ShowCommand struct {
}
type PassCommand struct {
}

type CleanCommand struct {
}

type PopCommand struct {
}

type HelpCommand struct {
}
type AckCommand struct {
}
