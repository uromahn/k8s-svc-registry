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
)

const (
	address = "localhost:9080"
	test    = "test"
)

var (
	op = flag.String("op", "r", "operation: r=register, u=unregister")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
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
		client := registry.NewServiceRegistryClient(conn, time.Duration(10)*time.Second)
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
