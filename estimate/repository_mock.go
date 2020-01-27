package estimate

type RepositoryMock struct {
	estimate Estimate
}

func (r *RepositoryMock) Get() (Estimate, error) {
	return r.estimate, nil
}

func (r *RepositoryMock) Save(estimate Estimate) error {
	r.estimate = estimate
	return nil
}
