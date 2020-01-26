package estimate

import (
	"time"
)

type RepositoryMock struct {
	estimate time.Duration
}

func (r *RepositoryMock) Get() (time.Duration, error) {
	return r.estimate, nil
}

func (r *RepositoryMock) Save(estimate time.Duration) error {
	r.estimate = estimate
	return nil
}
