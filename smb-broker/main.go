package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	var errChan = make(chan error)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", BrokerHandler(nil))
		errChan <- err
	}()
	fmt.Println("Started")
	err := <-errChan
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to start server: %v", err))
		os.Exit(1)
	}
}
