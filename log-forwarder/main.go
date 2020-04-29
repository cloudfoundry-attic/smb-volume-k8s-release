package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func main() {
	flag.Parse()
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "events", "cf-workloads",
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		time.Second * 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("event added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("event deleted: %s \n", obj)
			},
			UpdateFunc:func(oldObj, newObj interface{}) {
				fmt.Printf("event changed.  Old obj: %s, New obj: %s \n", oldObj, newObj)
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	for{
		time.Sleep(time.Second)
	}
}