package main

import (
	"flag"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"k8s.io/klog/v2"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"

	epwatcher "github.com/uromahn/k8s-svc-registry/pkg/endpointswatcher"
	kclient "github.com/uromahn/k8s-svc-registry/pkg/kubeclient"
	worker "github.com/uromahn/k8s-svc-registry/pkg/registrationworker"
	registry "github.com/uromahn/k8s-svc-registry/pkg/registry/server"
)

const (
	port            = ":9080"
	defaultLogLevel = "2"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             1 * time.Second, // If a client pings more than once every 1 second, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	Time:    1 * time.Second, // Ping the client if it is idle for 1 second to ensure the connection is still active
	Timeout: 1 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
}

var registrationQueue workqueue.RateLimitingInterface

func main() {
	klog.InitFlags(nil)
	flag.Set("v", defaultLogLevel)

	// initialize our Kubernetes client
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	k8sClient, err := kclient.NewKubeClient(kubeconfig, nil)
	if err != nil {
		klog.Fatalf("FATAL: cannot initialize Kubernetes client: %s", err.Error())
	}

	klog.Info("Creating IndexInformer for endpoints objects in all namespaces")
	indexInformer := epwatcher.CreateIndexInformer(k8sClient)
	stop := make(chan struct{})
	defer close(stop)

	klog.Info("Starting IndexInformer")
	go (*indexInformer).Run(stop)

	klog.Info("Waiting for endpoints cache to synchronized")
	syncError := false
	if !cache.WaitForCacheSync(stop, (*indexInformer).HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		syncError = true
	}

	klog.Info("Creating workqueue to process new service registrations")
	registrationQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	registry.InitRegistryServer(registrationQueue, k8sClient)

	registrationWorker := worker.NewWorker(registrationQueue, (*indexInformer).GetIndexer(), k8sClient)
	go (*registrationWorker).Run(stop)

	if !syncError {
		klog.Infof("listening for requests on localhost%s ...\n", port)
		lis, err := net.Listen("tcp", port)
		if err != nil {
			klog.Fatalf("failed to listen : %v", err)
		}
		s := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
		reg.RegisterServiceRegistryServer(s, &registry.ServiceRegistryServer{})
		if err := s.Serve(lis); err != nil {
			klog.Fatalf("failed to serve: %v", err)
		}
	} else {
		klog.Fatal("Cache sync error unrecoverable - exiting!")
	}
}
