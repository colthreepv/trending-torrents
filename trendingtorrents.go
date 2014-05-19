package main

import (
	"./fetchers"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	MINWAIT = 5 * time.Second
)

const (
	HTTPERROR = iota
)

type fetch struct {
	t0       time.Time
	duration time.Duration
}

func NewFetch() *fetch {
	return &fetch{t0: time.Now()}
}

func (l *fetch) calc() *fetch {
	l.duration = time.Now().Sub(l.t0)
	return l
}

// struct that contains N fetches
type fetchHistory struct {
	h        []*fetch
	quantity uint
}

func (f *fetchHistory) add(singleFetch *fetch) {
	f.h[f.quantity%uint(len(f.h))] = singleFetch
	f.quantity++
}

func (f *fetchHistory) String() string {
	var medianDuration time.Duration
	for _, v := range f.h {
		medianDuration += v.duration
	}
	medianDuration = medianDuration / time.Duration(len(f.h))
	return medianDuration.String()
}

func NewHistory(howmany int) *fetchHistory {
	return &fetchHistory{h: make([]*fetch, howmany)}
}

func fetchPage(hClient *http.Client, hChannel chan *http.Client, history *fetchHistory) (err error) {
	f := NewFetch() // start a timer
	page, err := hClient.Get("http://kickass.to/new/")
	if err != nil {
		return
	}
	defer page.Body.Close()
	if page.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http returned non-OK statuscode: %d\n", page.StatusCode))
	}

	// checks done, let's kaboom this HTML
	parsedPage, err := html.Parse(page.Body)
	if err != nil {
		return
	}
	// create a selector for kat
	katSelector := cascadia.MustCompile("table .odd, table .even")
	torrentElements := katSelector.MatchAll(parsedPage)

	katSlice := make([]*fetchers.KatRow, len(torrentElements))

	for index, element := range torrentElements {
		katSlice[index], err = fetchers.NewKatRow(element)
		if err != nil {
			fmt.Printf("trying creating a KatRow failed, :sadface:\n")
			continue
		}
		// uber debug
		// fmt.Printf("element: %v\n", katSlice[index])
	}

	// before returning the function "gives back" the Client to the channel
	hChannel <- hClient
	// enqueue history
	history.add(f.calc())
	return
}

type SpiderChannel struct {
	c   chan *http.Client
	qty uint16
}

func NewSpiderChan(howmany uint16) *SpiderChannel {
	return &SpiderChannel{c: make(chan *http.Client), qty: howmany}
}

func createHttpChannels(howmany int, channel chan *http.Client) (err error) {
	for i := 0; i < howmany; i++ {
		channel <- &http.Client{}
	}
	return nil
}

func main() {
	// sChannel := NewSpiderChan(uint16(400))
	hChannel := make(chan *http.Client)
	history := NewHistory(10)

	go createHttpChannels(4, hChannel)

	// gopher that handles the channel passing
	for {
		httpClient := <-hChannel
		if history.quantity > 20 {
			fmt.Printf("done 20 fetches, history: %s\n", history)
			break
		}
		// fmt.Printf("http client received, awesum: %#v\n", httpClient)
		go fetchPage(httpClient, hChannel, history)
	}

}
