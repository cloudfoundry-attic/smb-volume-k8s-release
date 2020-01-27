package main

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 2910))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	csi.RegisterIdentityServer(grpcServer, &noOpIdentityServer{})
	csi.RegisterNodeServer(grpcServer, &noOpNodeServer{})

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
