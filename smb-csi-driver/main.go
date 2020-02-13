package main

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/smb-csi-driver/identityserver"
	"code.cloudfoundry.org/smb-csi-driver/nodeserver"
	"flag"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	var endpoint = flag.String("endpoint", "", "")
	var nodeId = flag.String("nodeid", "", "")
	flag.Parse()

	println(*nodeId)

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

	println(">>>>>")
	println(proto)
	println(addr)
	println("<<<<<")

	lis, err := net.Listen(proto, addr)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	csi.RegisterIdentityServer(grpcServer, identityserver.NewSmbIdentityServer())
	csi.RegisterNodeServer(grpcServer, nodeserver.NewNodeServer(&execshim.ExecShim{}, &osshim.OsShim{}))

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
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
