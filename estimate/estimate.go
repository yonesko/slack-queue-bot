package estimate

import "time"

type Repository interface {
	Get() (time.Duration, error)
	Save(estimate time.Duration) error
}
