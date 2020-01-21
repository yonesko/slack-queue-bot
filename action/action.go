package action

type AddToQueue interface {
	Do(userId string) error
}
