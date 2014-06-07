package fetchers

import (
	"github.com/PuerkitoBio/goquery"

	// "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type KatRow struct {
	Name   string
	Magnet string
	Size   uint64
	Files  uint64
	Age    *time.Time
}

// helper function for NewKatPage
func parseSize(qty string) (uint64, error) {
	splitSize := strings.Fields(qty)
	amount, err := strconv.ParseFloat(strings.TrimSpace(splitSize[0]), 64)
	if err != nil {
		return 0, err
	}

	switch splitSize[1] {
	case "bytes":
		return uint64(amount), nil
	case "KB":
		return uint64(amount * 1024), nil
	case "MB":
		return uint64(amount * 1024 * 1024), nil
	case "GB":
		return uint64(amount * 1024 * 1024 * 1024), nil
	default:
		return 0, errors.New(fmt.Sprintf("quantity inside html is not recognized: %v\n", qty))
	}
}

// helper function for NewKatPage
func parseAge(timeAgo string) (*time.Time, error) {
	splitTime := strings.Fields(timeAgo)
	numericTime, err := strconv.ParseUint(splitTime[0], 10, 16)
	if err != nil {
		return nil, err
	}

	var durationAgo time.Duration
	switch splitTime[1] {
	case "sec.":
		durationAgo = time.Duration(numericTime) * time.Second
	case "min.":
		durationAgo = time.Duration(numericTime) * time.Minute
	case "hour", "hours":
		durationAgo = time.Duration(numericTime) * time.Hour
	case "day":
		durationAgo = time.Duration(numericTime) * time.Hour * 24
	case "week":
		durationAgo = time.Duration(numericTime) * time.Hour * 24 * 7
	default:
		return nil, errors.New(fmt.Sprintf("duration not handled: %v\n", splitTime[1]))
	}
	fromNow := time.Now().Add(-durationAgo)

	// uber debug
	// fmt.Printf("time input: %+v\ttime output: %+v\n", timeAgo, fromNow)
	return &fromNow, err
}

func NewKatPage(page *http.Response) ([]*KatRow, error) {
	doc, err := goquery.NewDocumentFromResponse(page)
	if err != nil {
		return nil, err
	}

	howManyTorrents := doc.Find("table .odd, table .even")
	torrentList := make([]*KatRow, howManyTorrents.Length())

	howManyTorrents.Each(func(i int, torrent *goquery.Selection) {
		name := torrent.Find(".cellMainLink").First().Text()
		magnet, _ := torrent.Find(".imagnet").First().Attr("href")
		size, err := parseSize(torrent.Find(".nobr.center").First().Text())
		if err != nil {
			fmt.Println(err)
			return
		}
		files, err := strconv.ParseUint(torrent.Find("td.center:not(.nobr)").First().Text(), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		age, err := parseAge(torrent.Find("td.center:not(.nobr) + td").First().Text())
		if err != nil {
			fmt.Println(err)
			return
		}
		torrentList[i] = &KatRow{Name: name, Magnet: magnet, Size: size, Files: files, Age: age}
	})
	// fmt.Printf("debug torrents:\n%#v\n", torrentList[4])
	return torrentList, nil
}

// helper function for KatScout
func fetchWithRetry(url string) (*http.Response, error) {
	i := 0
	for {
		htmlPage, err := http.Get(url)
		switch {
		case err != nil && i < 5:
			i++
			fmt.Println("kickass.to not responding, retrying in 5 seconds")
			time.Sleep(5000)
			continue
		case err != nil && i > 5:
			return nil, errors.New(fmt.Sprintf("too many errors, aborting the Scout: %s\n", err))
		case err == nil:
			return htmlPage, nil
		}
	}
}

func KatScout(done chan uint16) (err error) {
	htmlPage, err := fetchWithRetry("http://kickass.to/new/")
	if err != nil {
		fmt.Println(err)
		return
	}
	if htmlPage.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http returned non-OK statuscode: %d\n", htmlPage.StatusCode))
	}

	doc, err := goquery.NewDocumentFromResponse(htmlPage)
	if err != nil {
		return
	}

	pages, err := strconv.ParseUint(doc.Find(".turnoverButton.siteButton.bigButton:last-child").Text(), 10, 16)
	if err != nil {
		return
	}

	done <- uint16(pages)
	return nil
}

/**
 * Fetcher v2 Declaration
 *
 * Every KatFetch contains 25 torrent rows (max)
 * For each category, the max amout of KatFetch(ers) is 400
 */
type KatFetch struct {
	StartTime time.Time
	Elapsed   time.Duration
	success   bool `json:"-"`
	Data      []*KatRow
	fetchErr  error `json:"-"`
}

func (k *KatFetch) Fetch(httpClient *http.Client, httpChannel chan *http.Client, collection *KatFetchCollection) {
	page, err := collection.GetPage()
	if err != nil {
		if collection.ActiveFetchers == 0 {
			close(httpChannel)
		}
		// waste the http.Client here, since we're done!
		return
	}

	k.StartTime = time.Now()
	htmlPage, err := httpClient.Get(fmt.Sprintf("http://kickass.to/new/%d", page))
	fmt.Printf("requesting page: %s\n", fmt.Sprintf("http://kickass.to/new/%d", page))
	if err != nil {
		k.Fail(err, httpClient, httpChannel, collection, page)
		return
	}
	// defer htmlPage.Body.Close() goquery does that!!
	if htmlPage.StatusCode != http.StatusOK {
		k.Fail(
			errors.New(fmt.Sprintf("unexpected statusCode: %d instead of 200", htmlPage.StatusCode)),
			httpClient,
			httpChannel,
			collection,
			page)
		return
	}

	katSlice, err := NewKatPage(htmlPage)
	if err != nil {
		k.Fail(err, httpClient, httpChannel, collection, page)
		return
	}
	k.Done(katSlice, httpClient, httpChannel, collection)
}

func (k *KatFetch) Done(data []*KatRow, httpClient *http.Client, httpChannel chan *http.Client, collection *KatFetchCollection) {
	k.Elapsed = time.Now().Sub(k.StartTime)
	k.success = true
	k.Data = data
	collection.Completed <- k
	httpChannel <- httpClient
}

// one of the possible endings for the fetch operation
func (k *KatFetch) Fail(err error, httpClient *http.Client, httpChannel chan *http.Client, collection *KatFetchCollection, page int) {
	k.Elapsed = time.Now().Sub(k.StartTime)
	k.success = false
	k.fetchErr = err
	collection.ReturnPage(page)
	collection.Completed <- k
	httpChannel <- httpClient
	fmt.Printf("a fetcher failed: %s\n", err)
}

func NewKatFetch() *KatFetch {
	return &KatFetch{Data: make([]*KatRow, 25)}
}

/**
 * structure to export a KatFetchCollection for CouchDB
 */
type bulkExport struct {
	Docs []*KatFetch `json:"docs"`
}

/**
 * A KatFetch Collection contains ~howmany KatFetch(es)
 */
type KatFetchCollection struct {
	ActiveFetchers int `json:"-"`
	Current        int `json:"-"`

	Data     []*KatFetch `json:"successFetch"`
	Failures []*KatFetch `json:"failedFetch"`

	Completed chan *KatFetch `json:"-"`

	Board []bool `json:"-"`
}

func NewKatFetchCollection(howmany int) *KatFetchCollection {
	return &KatFetchCollection{
		Data:      make([]*KatFetch, howmany),
		Completed: make(chan *KatFetch),
		Board:     make([]bool, howmany),
		Failures:  make([]*KatFetch, 0),
	}
}

// ReceiveData is a function that should be called as goroutine, waiting for data to be sent
func (k *KatFetchCollection) ReceiveData() {
	for {
		newFetch, ok := <-k.Completed
		if ok == false {
			break
		}
		switch newFetch.success {
		case true:
			if err := k.Success(newFetch); err != nil {
				fmt.Println(err)
				close(k.Completed)
				break
			}
		case false:
			k.Failure(newFetch)
		}
	}
}

func (k *KatFetchCollection) Success(fetch *KatFetch) error {
	if k.Current == len(k.Data) {
		return errors.New(fmt.Sprintf("added too many fetches to a collection long only: %d\n", k.Current))
	}
	k.Data[k.Current] = fetch
	k.Current++
	k.ActiveFetchers--
	if k.ActiveFetchers < 9 {
		fmt.Printf("active fetchers: %d\n", k.ActiveFetchers)
	}
	return nil
}

func (k *KatFetchCollection) Failure(fetch *KatFetch) {
	k.Failures = append(k.Failures, fetch)
}

func (k *KatFetchCollection) Done() {
	// send k.Data to couchDB now!
	fmt.Println("all things received, ReceiveData shutting down!")
	close(k.Completed)
}

func (k *KatFetchCollection) GetPage() (int, error) {
	for i, _ := range k.Board {
		if k.Board[i] == false {
			k.Board[i] = true
			k.ActiveFetchers++
			if k.ActiveFetchers < 9 {
				fmt.Printf("active fetchers: %d\n", k.ActiveFetchers)
			}
			return i + 1, nil
		}
	}
	return 0, errors.New("board empty")
}

func (k *KatFetchCollection) ReturnPage(page int) {
	k.Board[page] = false
	k.ActiveFetchers--
}

func (k *KatFetchCollection) ExportSuccess() ([]byte, error) {
	docs := &bulkExport{Docs: k.Data}
	return json.Marshal(docs)
}

func (k *KatFetchCollection) ExportFailure() ([]byte, error) {
	docs := &bulkExport{Docs: k.Data}
	return json.Marshal(docs)
}
