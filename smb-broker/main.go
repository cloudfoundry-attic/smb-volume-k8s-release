package main

import (
	"code.cloudfoundry.org/smb-broker/store"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var errChan = make(chan error)
	go func() {
		handler, _ := BrokerHandler(&store.InMemoryServiceInstanceStore{})
		err := http.ListenAndServe("0.0.0.0:8080", handler)
		errChan <- err
	}()
	fmt.Println("Started")
	err := <-errChan
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to start server: %v", err))
		os.Exit(1)
	}
}
