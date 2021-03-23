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
