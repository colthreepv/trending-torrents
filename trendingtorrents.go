package main

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"net/http"
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
	name     string
	t0       time.Time
	duration time.Duration
}

func (l *lastFetch) start(name string) {
	l.name = name
	l.t0 = time.Now()
}

func (l *lastFetch) calc() *lastFetch {
	l.duration = time.Now().Sub(l.t0)
	return l
}

func (l *lastFetch) String() string {
	return fmt.Sprintf("timer %s: %s", l.name, l.duration)
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
		katSlice[index], err = fetchers.NewKatRow(element)
		if err != nil {
			fmt.Printf("trying creating a KatRow failed, :sadface:\n")
			continue
		}
		// uber debug
		fmt.Printf("element: %v\n", katSlice[index])
	}
	return
}

func main() {
	l := &lastFetch{}
	for {
		l.start("fetch Kat")
		err := fetchPage()
		if err != nil {
			fmt.Println(err)
			// os.Exit(HTTPERROR)
		}
		fmt.Println(l.calc())
		time.Sleep(MINWAIT)
	}
}
