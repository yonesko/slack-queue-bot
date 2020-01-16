package queue

import "testing"

func TestService_Add_DifferentUsers(t *testing.T) {
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
	_ = service.Add(User{Id: "ABC"})
	_ = service.Add(User{Id: "ABCD"})
	equals(queue, []string{"123", "ABC", "ABCD"})
}

func TestService_Pop(t *testing.T) {
	service := newInmemService()
	err := service.Pop()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(User{Id: "123"})
	if err != nil {
		t.Error(err)
	}
	err = service.Pop()
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{})
}

func TestService_DeleteAll(t *testing.T) {
	service := newInmemService()
	err := service.DeleteAll()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(User{Id: "123"})
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
