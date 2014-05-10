package main

import (
	"fmt"
	"github.com/ungerik/go-rss"
	// "net/http"
	"os"
	"time"
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

func fetchRSS() (err error) {
	// _, err = http.Get("http://kickass.to/new/rss/")
	rssChan, err := rss.Read("http://kickass.to/new/rss/")
	if err != nil {
		return err
	}
	for _, item := range rssChan.Item {
		fmt.Printf("item read: %s\n", item.Title)
	}
	return err
}

func main() {
	l := new(lastFetch)
	for {
		l.start()
		fetchErr := fetchRSS()
		if fetchErr != nil {
			os.Exit(HTTPERROR)
		}
		l.t1 = time.Now()
		fmt.Printf("time elapsed : %v\n", l.calc())
		time.Sleep(MINWAIT)
	}
}
