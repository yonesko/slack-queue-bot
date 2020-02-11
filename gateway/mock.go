package gateway

import "log"

type Mock struct {
}

func (m Mock) Send(userId, txt string) error {
	log.Printf("sending to %s '%s'", userId, txt)
	return nil
}

func (m Mock) SendAndLog(userId, txt string) {
	_ = m.Send(userId, txt)
}
