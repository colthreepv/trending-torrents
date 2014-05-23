package loggers

import (
	"fmt"
	"time"
)

type TimedRequest struct {
	t0       time.Time
	duration time.Duration
	success  bool
	err      error
}

func NewRequest() *TimedRequest {
	return &TimedRequest{t0: time.Now()}
}

func (t *TimedRequest) Done() *TimedRequest {
	t.duration = time.Now().Sub(t.t0)
	t.success = true
	return t
}

func (t *TimedRequest) Fail(err error) *TimedRequest {
	t.duration = time.Now().Sub(t.t0)
	t.err = err
	fmt.Println(err)
	return t
}

func (t *TimedRequest) String() string {
	return fmt.Sprint(t.duration)
}

// struct that contains N fetches
type RequestHistory struct {
	h        []*TimedRequest
	Quantity uint
}

func (f *RequestHistory) Add(singleFetch *TimedRequest) {
	f.h[f.Quantity%uint(len(f.h))] = singleFetch
	f.Quantity++
}

func (f *RequestHistory) String() string {
	var medianDuration time.Duration
	for _, v := range f.h {
		medianDuration += v.duration
	}
	medianDuration = medianDuration / time.Duration(len(f.h))
	return medianDuration.String()
}

func NewHistory(howmany int) *RequestHistory {
	return &RequestHistory{h: make([]*TimedRequest, howmany)}
}
