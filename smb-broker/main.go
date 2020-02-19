package main

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/smb-broker/handlers"
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
		namespace := os.Getenv("TARGET_NAMESPACE")
		handler, _ := handlers.BrokerHandler(clientset.CoreV1().PersistentVolumes(), clientset.CoreV1().PersistentVolumeClaims(namespace))
		err := http.ListenAndServe("0.0.0.0:8080", handler)
		errChan <- err
	}()
	logger := lager.NewLogger("smb-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	logger.Debug("Started")
	err = <-errChan
	if err != nil {
		logger.Error("Unable to start server", err)
		os.Exit(1)
	}
}
