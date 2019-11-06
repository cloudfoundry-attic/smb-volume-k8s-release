package main

import (
	"fmt"
	"net/http"
)

func main() {
	var errChan = make(chan error)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", BrokerHandler())
		errChan <- err
	}()
	fmt.Println("Started")
	<-errChan
}