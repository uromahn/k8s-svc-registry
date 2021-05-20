package main

import (
	"context"
	"flag"
	"log"
	"strconv"
	"time"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"
	registry "github.com/uromahn/k8s-svc-registry/pkg/registry/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	address = "localhost:9080"
	test    = "test"
)

var (
	op = flag.String("op", "r", "operation: r=register, u=unregister")
)

var kacp = keepalive.ClientParameters{
	Time:                2 * time.Second, // send pings every 2 seconds if there is no activity
	Timeout:             1 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,            // send pings even without active streams
}

/*
   =================================================================================
   This is just a simple test client. The actual client should do the following:
   - a go func that implements a health-check against an endpoint. It should support
     the following health-checks:
	 - tcp-check: will attempt to open a socket on ip-address:port and if it succeeds will close
	   the socket again and report success
	 - http-check: will attempt to issue a HTTP GET request against a supplied URL.
	   The health-check is considered successful if the response is a non 5xx code.
	 - file-check: will check for the presence of a file.
	   The health-check is considered successful if the file exists.
	 - future expansion: Unix domain socket check. This assumes that the server
	   implements a health endpoint and exposes it as a Unix domain socket.
	   The request will be a text command "ping" and the response is expected to be
	   "pong".
   - a pool of go functions with one to many gRPC clients connecting to one or many
     registration services.
   - the client will be configured using a YAML or JSON file with a specific set
     of configuration data which will be TBD.
   =================================================================================
*/
func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	var namedPorts []*reg.NamedPort

	port := reg.NamedPort{
		Port: 9080,
		Name: "http",
	}
	namedPorts = append(namedPorts, &port)

	client := registry.NewServiceRegistryClient(conn, time.Duration(10)*time.Second)
	for ip := 1; ip < 11; ip++ {
		ipAddr := "192.168.1." + strconv.Itoa(ip)
		svcInfo := &reg.ServiceInfo{
			Namespace:   "test-dev",
			ServiceName: "test",
			HostName:    "localhost",
			Ipaddress:   ipAddr,
			NodeName:    "uromahn-vm-ubuntu18",
			Ports:       namedPorts,
			Weight:      1.0,
		}
		ctx := context.Background()

		// contact the server with a registration message and print the result
		if *op == "r" {
			regResult, err := client.Register(ctx, svcInfo)
			if err != nil {
				log.Fatalf("could not register service: %v", err)
			}
			log.Printf("Registered: %v", regResult)
		} else if *op == "u" {
			unregResult, err := client.UnRegister(ctx, svcInfo)
			if err != nil {
				log.Fatalf("could not unregister service: %v", err)
			}
			log.Printf("Unregistered: %v", unregResult)
		} else {
			log.Printf("Invalid operation: %s", *op)
		}
	}
}
