package main

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/mrgamer/trendingtorrents/fetchers"
)

const (
	MINWAIT = 5 * time.Second
)

const (
	HTTPERROR = iota
)

type lastFetch struct {
	t0, t1  time.Time
	lastRun time.Duration
}

func (lst *lastFetch) start() {
	lst.t0 = time.Now()
}

func (lst *lastFetch) calc() time.Duration {
	lst.lastRun = lst.t1.Sub(lst.t0)
	return lst.lastRun
}

func fetchPage() (err error) {
	page, err := http.Get("http://kickass.to/new/")
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

	katSlice := make([] *fetchers.KatRow, len(torrentElements))

	for index, element := range torrentElements {
		katSlice[index] = fetchers.NewKatRow(element)
		fmt.Printf("element: %v\n", katSlice[index])
	}
	return
}

func main() {
	l := new(lastFetch)
	for {
		l.start()
		fetchErr := fetchPage()
		if fetchErr != nil {
			os.Exit(HTTPERROR)
		}
		l.t1 = time.Now()
		fmt.Printf("time elapsed : %v\n", l.calc())
		time.Sleep(MINWAIT)
	}
}
