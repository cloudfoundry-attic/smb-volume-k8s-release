package main

import (
	"fmt"
	"net/http"
)

func main() {
	var errChan = make(chan error)
	go func() {
		err := http.ListenAndServe("localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("Hi"))
			if err != nil {
				panic(err)
			}
		}))
		errChan <- err
	}()
	fmt.Println("Started")
	<-errChan
}