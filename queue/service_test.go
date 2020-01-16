package queue

import "testing"

func TestService_Add_1(t *testing.T) {
	service := newInmemService()
	err := service.Add(User{Id: "123"})
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{"123"})
}

func TestService_Add_Idempotent(t *testing.T) {
	service := newInmemService()
	err := service.Add(User{Id: "123"})
	if err != nil {
		t.Error(err)
	}
	err = service.Add(User{Id: "123"})
	if err == nil || err.Error() != "already exist" {
		t.Error("must be already exist")
	}
}
