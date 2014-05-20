package loggers

import (
	"time"
)

type TimedRequest struct {
	t0       time.Time
	duration time.Duration
}

func NewRequest() *TimedRequest {
	return &TimedRequest{t0: time.Now()}
}

func (l *TimedRequest) Calc() *TimedRequest {
	l.duration = time.Now().Sub(l.t0)
	return l
}

// struct that contains N fetches
type FetchHistory struct {
	h        []*TimedRequest
	Quantity uint
}

func (f *FetchHistory) Add(singleFetch *TimedRequest) {
	f.h[f.Quantity%uint(len(f.h))] = singleFetch
	f.Quantity++
}

func (f *FetchHistory) String() string {
	var medianDuration time.Duration
	for _, v := range f.h {
		medianDuration += v.duration
	}
	medianDuration = medianDuration / time.Duration(len(f.h))
	return medianDuration.String()
}

func NewHistory(howmany int) *FetchHistory {
	return &FetchHistory{h: make([]*TimedRequest, howmany)}
}
