package main

import (
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/volume-services-log-forwarder/filter"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder"
	"code.cloudfoundry.org/volume-services-log-forwarder/watcher"
	"flag"
	"github.com/fluent/fluent-logger-golang/fluent"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"log"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func main() {
	flag.Parse()
	config, err := rest.InClusterConfig()
	log.Printf("config: %#v", config)
	if err != nil {
		panic(err.Error())
	}

	logger, _ := lagerflags.NewFromConfig("log-forwarder", lagerflags.LagerConfig{LogLevel:lagerflags.DEBUG})

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// instantiate watcher (filter, forwarder, fluent)
	fluentConfig := fluent.Config{
		FluentHost: "fluentd-forwarder-ingress.cf-system.svc.cluster.local",
	}
	fluentClient, _ := fluent.New(fluentConfig)
	fltr := filter.NewFilter(logger)
	fwdr := forwarder.NewForwarder(logger, fluentClient)
	podInterface := clientset.CoreV1().Pods("cf-workloads")
	wtchr := watcher.NewWatcher(logger, fltr, fwdr, podInterface)

	logger.Debug("creating watchlist")
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "events", "cf-workloads",
		fields.Everything())
	logger.Debug("created watchlist")

	logger.Debug("creating informer")
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		time.Second * 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: wtchr.Watch,
			DeleteFunc: func(obj interface{}) {
				//noop
			},
			UpdateFunc:func(oldObj, newObj interface{}) {
				//noop
			},
		},
	)
	logger.Debug("created informer")

	logger.Debug("watching...")
	stop := make(chan struct{})
	go controller.Run(stop)
	for{
		time.Sleep(time.Second)
	}
	logger.Debug("done")
}