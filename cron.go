package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type couchGet []string

func main() {
	listFetches, err := http.Get("http://localhost:5984/trendingtorrents/_design/convert/_list/listFetches/getFetches")
	if err != nil {
		fmt.Println(err)
	}
	fetches, err := ioutil.ReadAll(listFetches.Body)
	if err != nil {
		fmt.Println(err)
	}
	var response []string
	err = json.Unmarshal(fetches, &response)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range response {
		_, err := http.Post(fmt.Sprintf("http://localhost:5984/trendingtorrents/_design/convert/_update/convert/%s", v), "application/json", nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}
