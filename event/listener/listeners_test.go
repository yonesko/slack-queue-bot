package listener

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/estimate"
	"github.com/yonesko/slack-queue-bot/model"
	"strconv"
	"testing"
	"time"
)

func TestHoldTimeEstimateListener_FirstInQueue(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "123",
		PrevHolderUserId:    "",
		AuthorUserId:        "123",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "abc",
		PrevHolderUserId:    "123",
		AuthorUserId:        "123",
		Ts:                  time.Unix(int64((time.Minute * 35).Seconds()), 0),
	})
	duration, err := rep.Read()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Minute * 35, 1}, duration)
}

func TestHoldTimeEstimateListener_TooLongTime(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "123",
		PrevHolderUserId:    "",
		AuthorUserId:        "123",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "abc",
		PrevHolderUserId:    "123",
		AuthorUserId:        "123",
		Ts:                  time.Unix(int64((time.Hour*2).Seconds())+1, 0),
	})
	duration, err := rep.Read()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{0, 0}, duration)
}

func TestHoldTimeEstimateListener_InMiddleOfQueue(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "1",
		Ts:                  time.Unix(int64((time.Minute * 35).Seconds()), 0),
	})
	duration, err := rep.Read()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Minute * 35, 1}, duration)
}

func TestHoldTimeEstimateListener_ForceDel(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(model.NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "4",
		Ts:                  time.Unix(int64((time.Minute * 35).Seconds()), 0),
	})
	duration, err := rep.Read()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Second * 0, 0}, duration)
}

func TestHoldTimeEstimateListener_MultiplyEvents(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	for i := 1; i <= 100; i++ {
		listener.Fire(model.NewHolderEvent{
			CurrentHolderUserId: strconv.Itoa(i),
			PrevHolderUserId:    strconv.Itoa(i - 1),
			AuthorUserId:        strconv.Itoa(i - 1),
			Ts:                  time.Unix(int64(i)*int64((time.Minute*35).Seconds()), 0),
		})
	}

	duration, err := rep.Read()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Minute * 35, 99}, duration)
}
