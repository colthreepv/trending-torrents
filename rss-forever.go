package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"code.google.com/p/go.net/html"
	"net/http"
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

type Torrent struct {
	OrientClass string `json:"@class"`
	Category []string `json:"category"`
	Description string `json:"description"`
	Hash string `json:"hash"`
	Link string `json:"link"`
	PubDate time.Time `json:"pubDate"`
	Title string `json:"title"`
}

func NewTorrentFromRss (i rss.Item) *Torrent, error {
	ti := new(Torrent)
	ti.Title = i.Title
	ti.Category = 
}

func fetchRSS() error {
	list, listErr := http.Get("http://kickass.to/new/")
	// rssChan, err := rss.Read("http://kickass.to/new/")
	if listErr != nil {
		return err
	}
	for _, item := range rssChan.Item {

		fmt.Printf("item read: %s\n", item.Title)
		/* prepare RSS-XML to JSON
		jsonifiedItem, jErr := json.Marshal(item)
		if jErr != nil {
			return jErr
		}

		// build request
		req, hErr := http.NewRequest("POST", "http://localhost:2480/document/trending/", bytes.NewReader(jsonifiedItem))
		if hErr != nil {
			return hErr
		}
		req.SetBasicAuth("root", "brolazzo0")
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		fmt.Printf("debug request: %s\n", jsonifiedItem)
		// send it
		resp, rErr := http.DefaultClient.Do(req)
		if rErr != nil {
			return rErr
		}
		fmt.Printf("response from DB: %v\n", resp.StatusCode)
		*/
		return nil // debug, stop only after first one
	}
	return nil
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
