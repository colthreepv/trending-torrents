package fetchers

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type KatRow struct {
	Name   string
	Magnet string
	Size   uint64
Files  uint32
	Age    time.Time
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

func NewKatRow(n *html.Node) (k *KatRow) {
	nameSelector := cascadia.MustCompile(".cellMainLink")
	name := nameSelector.MatchFirst(n).FirstChild.Data

	magnetSelector := cascadia.MustCompile(".imagnet")
	var magnet string
	for _, attribute := range magnetSelector.MatchFirst(n).Attr {
		if attribute.Key == "href" {
			magnet = attribute.Val
			break
		}
	}

	sizeSelector := cascadia.MustCompile(".nobr.center")
	sizeEl := sizeSelector.MatchFirst(n).FirstChild
	sizeQty := sizeEl.NextSibling.FirstChild.Data

	sizeAmount , err := strconv.ParseFloat(sizeEl.Data, 32)
	if err != nil {
		fmt.Printf("trying to parse: %#v", sizeEl.Data)
		return
	}
	size, err := parseSize(float32(sizeAmount), sizeQty)
	if err != nil {
		fmt.Println("Error sizesss!!!")
		return nil
	}

	return &KatRow{Name: name, Magnet: magnet, Size: size}
}
