package estimate

import (
	"time"
)

type Estimate struct {
	Average     time.Duration
	Estimations int
}

func (e Estimate) AddOne(duration time.Duration) Estimate {
	sum := e.Average.Milliseconds()*int64(e.Estimations) + duration.Milliseconds()
	return Estimate{
		Average:     time.Duration(time.Millisecond.Nanoseconds() * (sum / (int64(e.Estimations) + 1))),
		Estimations: e.Estimations + 1,
	}
}

type Repository interface {
	Get() (Estimate, error)
	Save(estimate Estimate) error
}
