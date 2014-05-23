package main

import (
	"github.com/mrgamer/trendingtorrents/fetchers"
	"github.com/mrgamer/trendingtorrents/loggers"

	"encoding/json"
	"fmt"
	"io/ioutil"
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
	overAllTime := loggers.NewRequest()
	hChannel := make(chan *http.Client)
	KatReady := make(chan uint16)

	go createHttpChannels(4, hChannel)
	go fetchers.KatScout(KatReady)

	// this will become a gopher
	pages := <-KatReady
	fmt.Printf("pages reported from scout: %d\n", pages)
	kCollection := fetchers.NewKatFetchCollection(int(100))

	go kCollection.ReceiveData()
	// maybe board inside collection?
	// board := &Board{Items: make([]bool, int(16))}

	// function spec
	// func ()
	for {
		httpClient, ok := <-hChannel
		if ok == false {
			fmt.Println("all is done!")
			kJSON, err := json.Marshal(kCollection)
			if err != nil {
				fmt.Println(err)
			}
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
