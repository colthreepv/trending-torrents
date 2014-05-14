package fetchers

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"time"
)

type KatRow struct {
	Name   string
	Magnet string
	Size   float32
	Files  uint32
	Age    time.Time
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
	return &KatRow{Name: name, Magnet: magnet}
}
