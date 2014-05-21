package main

import (
	"github.com/mrgamer/trendingtorrents/fetchers"
	"github.com/mrgamer/trendingtorrents/loggers"

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

func fetchPage(hClient *http.Client, hChannel chan *http.Client, history *loggers.RequestHistory, b *Board) (err error) {
	f := loggers.NewRequest() // start a timer
	p, err := b.Get()
	if err != nil { // pages ended!
		return
	}
	htmlPage, err := hClient.Get(fmt.Sprintf("http://kickass.to/new/%d", p))
	fmt.Printf("requesting page: %s\n", fmt.Sprintf("http://kickass.to/new/%d", p))
	if err != nil {
		history.Add(f.Fail(err)) // log errors to orient
		b.Fail(p)
		return
	}
	defer htmlPage.Body.Close()
	if htmlPage.StatusCode != http.StatusOK {
		history.Add(f.Fail(err)) // log errors to orient
		b.Fail(p)
		return errors.New(fmt.Sprintf("http returned non-OK statuscode: %d\n", htmlPage.StatusCode))
	}

	// checks done, let's kaboom this HTML
	parsedPage, err := html.Parse(htmlPage.Body)
	if err != nil {
		b.Fail(p)
		return
	}
	// create a selector for kat
	katSelector := cascadia.MustCompile("table .odd, table .even")
	torrentElements := katSelector.MatchAll(parsedPage)

	katSlice := make([]*fetchers.KatRow, len(torrentElements))

	for index, element := range torrentElements {
		katSlice[index], err = fetchers.NewKatRow(element)
		if err != nil {
			fmt.Printf("KatRow failed: %s\n", err)
			continue
		}
		// uber debug
		// fmt.Printf("element: %v\n", katSlice[index])
	}

	// before returning the function "gives back" the Client to the channel
	hChannel <- hClient
	// enqueue history
	history.Add(f.Done())
	return
}

func createHttpChannels(howmany int, channel chan *http.Client) (err error) {
	for i := 0; i < howmany; i++ {
		channel <- &http.Client{}
	}
	return nil
}

type Board struct {
	Items []bool
}

// return the first non-fetched page
func (b *Board) Get() (int, error) {
	for i, v := range b.Items {
		if v == false {
			b.Items[i] = true
			return i, nil
		}
	}
	return 0, errors.New("board empty")
}

// in case of fetch failure it gives me back the page
func (b *Board) Fail(page int) {
	b.Items[page] = false
}

func main() {
	hChannel := make(chan *http.Client)
	history := loggers.NewHistory(10)
	KatReady := make(chan uint16)

	go createHttpChannels(4, hChannel)
	go fetchers.KatScout(KatReady, history)

	// this will become a gopher
	pages := <-KatReady
	fmt.Printf("pages reported from scout: %d\n", pages)
	board := &Board{Items: make([]bool, int(16))}
	for {
		httpClient := <-hChannel
		if history.Quantity > 20 {
			fmt.Printf("done 20 fetches, history: %s\n", history)
			break
		}
		// fmt.Printf("http client received, awesum: %#v\n", httpClient)
		go fetchPage(httpClient, hChannel, history, board)
	}

}
