package gateway

type Mock struct {
}

func (m Mock) Send(userId, txt string) error {
	return nil
}

func (m Mock) SendAndLog(userId, txt string) {
	return
}
