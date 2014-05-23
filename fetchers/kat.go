package fetchers

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"

	"bytes"
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

func parseSize(amount float32, qty string) (uint64, error) {
	switch qty {
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

func NewKatRow(n *html.Node) (k *KatRow, err error) {
	name := cascadia.MustCompile(".cellMainLink").MatchFirst(n).FirstChild.Data

	// magnet parsing
	var magnet string
	for _, attribute := range cascadia.MustCompile(".imagnet").MatchFirst(n).Attr {
		if attribute.Key == "href" {
			magnet = attribute.Val
			break
		}
	}

	// size parsing
	statsEl := cascadia.MustCompile(".nobr.center").MatchFirst(n)
	sizeEl := statsEl.FirstChild
	sizeQty := sizeEl.NextSibling.FirstChild.Data

	sizeAmount, err := strconv.ParseFloat(strings.TrimSpace(sizeEl.Data), 32)
	if err != nil {
		return nil, err
	}
	size, err := parseSize(float32(sizeAmount), sizeQty)
	if err != nil {
		return nil, err
	}

	files, err := strconv.ParseUint(strings.TrimSpace(cascadia.MustCompile("td.center:not(.nobr)").MatchFirst(n).FirstChild.Data), 10, 64)
	if err != nil {
		return nil, err
	}

	age, err := parseAge(cascadia.MustCompile("td.center:not(.nobr) + td").MatchFirst(n).FirstChild.Data)
	if err != nil {
		return nil, err
	}

	return &KatRow{Name: name, Magnet: magnet, Size: size, Files: files, Age: age}, nil
}

func getPage(url string) (*http.Response, error) {
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
	htmlPage, err := getPage("http://kickass.to/new/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer htmlPage.Body.Close()
	if htmlPage.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http returned non-OK statuscode: %d\n", htmlPage.StatusCode))
	}

	// checks done, let's kaboom this HTML
	parsedPage, err := html.Parse(htmlPage.Body)
	if err != nil {
		return
	}

	// read how many pages
	pages := cascadia.MustCompile(".turnoverButton.siteButton.bigButton:last-child").MatchFirst(parsedPage).FirstChild.Data
	pagesParsed, err := strconv.ParseUint(pages, 10, 16)
	if err != nil {
		return
	}

	done <- uint16(pagesParsed)
	return nil
}

/**
 * Fetcher v2 Declaration
 *
 * Every KatFetch contains 25 torrent rows (max)
 * For each category, the max amout of KatFetch(ers) is 400
 */
type KatFetch struct {
	startTime time.Time
	Elapsed   time.Duration
	Success   bool
	Data      []*KatRow
	Err       error
}

func (k *KatFetch) Fetch(httpClient *http.Client, httpChannel chan *http.Client, collectionChannel chan *KatFetch, page uint16) {
	k.startTime = time.Now()
	htmlPage, err := httpClient.Get(fmt.Sprintf("http://kickass.to/new/%d", page))
	fmt.Printf("requesting page: %s\n", fmt.Sprintf("http://kickass.to/new/%d", page))
	if err != nil {
		k.Fail(err, httpClient, httpChannel)
		return
	}
	defer htmlPage.Body.Close()
	if htmlPage.StatusCode != http.StatusOK {
		k.Fail(err, httpClient, httpChannel)
		return
	}

	// checks done, let's kaboom this HTML
	parsedPage, err := html.Parse(htmlPage.Body)
	if err != nil {
		k.Fail(err, httpClient, httpChannel)
		return
	}
	// create a selector for kat
	katSelector := cascadia.MustCompile("table .odd, table .even")
	torrentElements := katSelector.MatchAll(parsedPage)

	katSlice := make([]*KatRow, len(torrentElements))

	for index, element := range torrentElements {
		katSlice[index], err = NewKatRow(element)
		if err != nil {
			fmt.Printf("KatRow failed: %s\n", err)
			continue
		}
		// uber debug
		// fmt.Printf("element: %v\n", katSlice[index])
	}

	k.Done(katSlice, httpClient, httpChannel, collectionChannel)
}

func (k *KatFetch) Done(data []*KatRow, httpClient *http.Client, httpChannel chan *http.Client, collectionChannel chan *KatFetch) {
	k.Success = true
	k.Data = data
	httpChannel <- httpClient
	collectionChannel <- k
}

// one of the possible endings for the fetch operation
func (k *KatFetch) Fail(err error, httpClient *http.Client, httpChannel chan *http.Client) {
	k.Elapsed = time.Now().Sub(k.startTime)
	k.Success = false
	k.Err = err
	httpChannel <- httpClient
	fmt.Printf("a fetcher failed: %s\n", err)
}

func NewKatFetch() *KatFetch {
	return &KatFetch{Data: make([]*KatRow, 25)}
}

/**
 * A KatFetch Collection contains ~howmany KatFetch(es)
 */
type KatFetchCollection struct {
	Data    []*KatFetch
	Length  int
	Channel chan *KatFetch `json:"-"`
}

// ReceiveData is a function that should be called as goroutine, waiting for data to be sent
func (k *KatFetchCollection) ReceiveData() {
	for {
		newFetch, ok := <-k.Channel
		if ok == false {
			break
		}
		k.Add(newFetch)
	}
}

func (k *KatFetchCollection) Add(fetch *KatFetch) {
	k.Data[k.Length%len(k.Data)] = fetch
	k.Length++
}

func (k *KatFetchCollection) Done() {
	// send data now!
	fmt.Println("all things received, ReceiveData shutting down!")
	close(k.Channel)
}

func (k *KatFetchCollection) MarshalJSON() ([]byte, error) {
	dataSlice := make([][]byte, len(k.Data))

	for i := 0; i < len(k.Data); i++ {
		dataJSON, err := json.Marshal(k.Data[i])
		if err != nil {
			return nil, err
		}
		switch i {
		case 0:
			dataSlice[i] = append([]byte("["), dataJSON...)
		case len(k.Data) - 1:
			dataSlice[i] = append(dataJSON, []byte("]")...)
		default:
			dataSlice[i] = dataJSON
		}
	}
	return bytes.Join(dataSlice, []byte(",")), nil
}

func NewKatFetchCollection(howmany int) *KatFetchCollection {
	return &KatFetchCollection{Data: make([]*KatFetch, howmany), Channel: make(chan *KatFetch)}
}
