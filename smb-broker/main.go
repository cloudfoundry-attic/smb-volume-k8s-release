package main

import (
	"code.cloudfoundry.org/smb-broker/handlers"
	"code.cloudfoundry.org/smb-broker/store"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var errChan = make(chan error)
	go func() {
		namespace := "eirini"
		handler, _ := handlers.BrokerHandler(&store.InMemoryServiceInstanceStore{}, clientset.CoreV1().PersistentVolumes(), clientset.CoreV1().PersistentVolumeClaims(namespace))
		err := http.ListenAndServe("0.0.0.0:8080", handler)
		errChan <- err
	}()
	fmt.Println("Started")
	err = <-errChan
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to start server: %v", err))
		os.Exit(1)
	}
}
