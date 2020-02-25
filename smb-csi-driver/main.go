package main

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/smb-csi-driver/identityserver"
	"code.cloudfoundry.org/smb-csi-driver/nodeserver"
	"code.cloudfoundry.org/lager"
	"flag"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"os"
	"strings"
)

type unaryInterceptor struct {
	logger lager.Logger
}

func main() {
	var endpoint = flag.String("endpoint", "", "")
	var nodeId = flag.String("nodeid", "", "")
	flag.Parse()

	logger := lager.NewLogger("smb-csi-driver")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	interceptor := unaryInterceptor{logger: logger}

	logger.Info(fmt.Sprintf("node-id: %s", *nodeId))

	proto, addr, err := ParseEndpoint(*endpoint)
	if err != nil {
		log.Fatal(err.Error())
	}

	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			log.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		}
	}

	logger.Info(">>>>>")
	logger.Info(proto)
	logger.Info(addr)
	logger.Info("<<<<<")

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	lis, err := net.Listen(proto, addr)

	if err != nil {
		logger.Fatal("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.logGRPC),
	}

	grpcServer := grpc.NewServer(opts...)
	csi.RegisterIdentityServer(grpcServer, identityserver.NewSmbIdentityServer())
	csi.RegisterNodeServer(grpcServer, nodeserver.NewNodeServer(logger, &execshim.ExecShim{}, &osshim.OsShim{}, clientset.CoreV1().ConfigMaps("default")))

	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Fatal("failed to serve", err, lager.Data{"listener": lis})
	}
}

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}

func (l unaryInterceptor) logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l.logger.Info("GRPC request", lager.Data{"method": info.FullMethod, "req": protosanitizer.StripSecrets(req).String()})
	resp, err := handler(ctx, req)
	if err != nil {
		l.logger.Error("GRPC error", err)
	} else {
		l.logger.Info("GRPC response", lager.Data{"response": protosanitizer.StripSecrets(resp).String()})
	}
	return resp, err
}