package main

import (
	"github.com/mrgamer/trendingtorrents/fetchers"
	"github.com/mrgamer/trendingtorrents/loggers"

	tt "github.com/jackpal/Taipei-Torrent/torrent"

	// "bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	MINWAIT = 5 * time.Second
)

const (
	HTTPERROR = iota
)

func createHttpChannels(howmany int, channel chan *http.Client) (err error) {
	for i := 0; i < howmany; i++ {
		channel <- &http.Client{}
	}
	return nil
}

func main() {
	tSession, tError := tt.NewTorrentSession(&tt.TorrentFlags{
		Dial:                nil,
		Port:                7777,
		FileDir:             ".",
		SeedRatio:           1,
		UseDeadlockDetector: false,
		UseLPD:              false,
		UseDHT:              true,
		UseUPnP:             true,
		UseNATPMP:           true,
		TrackerlessMode:     false,
		Gateway:             "",
	}, "magnet:?xt=urn:btih:680A1BE8D653E60DD6B4D1BEEA3702AC87A3756C&dn=into+the+storm+2014+1080p+brrip+x264+yify&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337", 7778)
	log.Printf("%v %v", tSession, tError)
	// tSession.
	overAllTime := loggers.NewRequest()
	hChannel := make(chan *http.Client)
	KatReady := make(chan uint16)

	go createHttpChannels(10, hChannel)
	go fetchers.KatScout(KatReady)

	// this will become a gopher
	pages := <-KatReady
	fmt.Printf("pages reported from scout: %d\n", pages)
	kCollection := fetchers.NewKatFetchCollection(int(1))

	go kCollection.ReceiveData()
	// maybe board inside collection?
	// board := &Board{Items: make([]bool, int(16))}

	// function spec
	// func ()
	for {
		httpClient, ok := <-hChannel
		if ok == false {
			fmt.Println("all is done!")
			kJSON, err := kCollection.ExportSuccess()
			if err != nil {
				fmt.Println(err)
			}
			// kJsonReader := bytes.NewReader(kJson)
			// // couchResponse, err := http.Post("http://localhost:5984/trendingtorrents/_bulk_docs", "application/json", kJsonReader)
			// // couchResponse.Body.Close()
			// // if err != nil {
			// // 	fmt.Println(err)
			// // }
			err = ioutil.WriteFile("katfetch.json", kJSON, 0644)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("produced katfetch.json")
			fmt.Printf("All done in: %s\n", overAllTime.Done())
			break
		}

		f := fetchers.NewKatFetch()
		go f.Fetch(httpClient, hChannel, kCollection)
	}

}
