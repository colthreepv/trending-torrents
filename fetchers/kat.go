package fetchers

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
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
		durationAgo = time.Duration(numericTime)*time.Second
	case "min.":
		durationAgo = time.Duration(numericTime)*time.Minute
	case "hours":
		durationAgo = time.Duration(numericTime)*time.Hour
	case "day":
		durationAgo = time.Duration(numericTime)*time.Hour*24
	case "week":
		durationAgo = time.Duration(numericTime)*time.Hour*24*7
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
