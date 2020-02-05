package estimate

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

func (e Estimate) TimeToWait(before uint, holdStart time.Time) time.Duration {
	if before <= 0 {
		return 0
	}
	return time.Duration(int64(before-1)*e.Average.Nanoseconds()) + e.avgRestOfHolding(holdStart)
}

func (e Estimate) avgRestOfHolding(holdStart time.Time) time.Duration {
	holdRest := e.Average - time.Now().Sub(holdStart)
	if holdRest.Nanoseconds() < 0 {
		return 0
	}
	return holdRest
}

type Repository interface {
	Read() (Estimate, error)
	Save(estimate Estimate) error
}

type fileRepository struct {
	filename string
}

func NewRepository() *fileRepository {
	createDbIfNeed()
	return &fileRepository{filename: "db/estimate.json"}
}
func createDbIfNeed() {
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		err := os.Mkdir("db", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}
func (f *fileRepository) Save(estimate Estimate) error {
	bytes, err := json.Marshal(estimate)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f.filename, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileRepository) Read() (Estimate, error) {
	bytes, err := ioutil.ReadFile(f.filename)
	if os.IsNotExist(err) {
		return Estimate{}, nil
	}
	if err != nil {
		return Estimate{}, err
	}
	queue := &Estimate{}
	err = json.Unmarshal(bytes, queue)
	if err != nil {
		return Estimate{}, err
	}
	return *queue, nil
}
