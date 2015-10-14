package goProxyServer

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type Client struct {
	http.Client
	Request *http.Request
	ClientNumber int32
}

var clientQueue = make(chan *Client, 2)
var clientNumber int32

func RequestContent(url string) (string, error) {
	client, err := getNextClient(url)
	if err != nil {
		log.Fatalf("Could not retrieve client. %s", err)
	}
	res, err := client.Do(client.Request)
	time.Sleep(100 * time.Millisecond)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	go enqueueClient(client)
	return string(body), nil
}

func getNextClient(url string) (*Client, error) {
	select {
	case client := <- clientQueue:
		log.Println("Reusing queued client.")
		return client, nil
	case <- time.After(100 * time.Millisecond):
		log.Println("Returning new client.")
		req, _ := http.NewRequest("GET", url, nil)
		return &Client{
			Request: req,
			ClientNumber: atomic.AddInt32(&clientNumber, 1),
		}, nil
	}
}

func enqueueClient(client *Client) {
	select {
	case clientQueue <- client:
		log.Printf("Client [%d] placed on queue for reuse.", client.ClientNumber)
	case <- time.After(2 * time.Second):
		log.Printf("Queue has been sitting full. Abandoning extra instance [%d].", client.ClientNumber)
	}
}
