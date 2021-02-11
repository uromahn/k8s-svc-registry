package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"path/filepath"

	"k8s.io/klog/v2"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"

	epwatcher "github.com/uromahn/k8s-svc-registry/internal/endpointswatcher"
	kclient "github.com/uromahn/k8s-svc-registry/internal/kubeclient"
	worker "github.com/uromahn/k8s-svc-registry/internal/registrationworker"
	servertypes "github.com/uromahn/k8s-svc-registry/internal/servertypes"
)

const (
	port = ":9080"
)

var registrationQueue workqueue.RateLimitingInterface

// register implements registry.ServiceRegistryService.Register
func register(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	ports := svcInfo.GetPorts()
	klog.Infof("Received registration request for namespace   = %s", svcInfo.GetNamespace())
	klog.Infof("                                  service     = %s", svcInfo.GetServiceName())
	klog.Infof("                                  hostname    = %s", svcInfo.GetHostName())
	klog.Infof("                                  ipAddress   = %s", svcInfo.GetIpaddress())
	for _, port := range ports {
		klog.Infof("                                  port        = %d", port.GetPort())
		klog.Infof("                                  portName    = %s", port.GetName())
	}

	// here we call our Kubernetes API to create a corresponding entry in the services endpoints object
	// make sure we have a namespace
	_, err := kclient.GetOrCreateNamespace(ctx, svcInfo.GetNamespace(), true)
	if err != nil {
		klog.Errorf("ERROR: could not get or create namespace '%s'", svcInfo.GetNamespace())
		return nil, err
	}
	// make sure we have the service object created
	_, err = kclient.GetOrCreateService(ctx, svcInfo.GetNamespace(), svcInfo.GetServiceName(), svcInfo.GetPorts(), true)
	if err != nil {
		klog.Errorf("ERROR: unable to get or create service '%s' in namespace '%s'", svcInfo.GetServiceName(), svcInfo.GetNamespace())
		return nil, err
	}

	// create the response channel
	respChannel := make(chan servertypes.ResultMsg)
	// and the registration message
	regMsg := servertypes.RegistrationMsg{
		Ctx:             ctx,
		ResponseChannel: respChannel,
		SvcInfo:         svcInfo,
		Op:              servertypes.Register,
	}
	// send it to the worker via our queue
	registrationQueue.Add(regMsg)
	// wait for the response
	respMsg := <-respChannel
	// we expect the worker to close the channel after the response has been sent
	return respMsg.Result, respMsg.Err
}

// unregister implements registry.ServiceREgistryService/UnRegister
func unRegister(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	ports := svcInfo.GetPorts()
	klog.Infof("Received unregistration request for namespace   = %s", svcInfo.GetNamespace())
	klog.Infof("                                    service     = %s", svcInfo.GetServiceName())
	klog.Infof("                                    hostname    = %s", svcInfo.GetHostName())
	klog.Infof("                                    ipAddress   = %s", svcInfo.GetIpaddress())
	for _, port := range ports {
		klog.Infof("                                    port        = %d", port.GetPort())
		klog.Infof("                                    portName    = %s", port.GetName())
	}

	// create the response channel
	respChannel := make(chan servertypes.ResultMsg)
	// and the registration message
	regMsg := servertypes.RegistrationMsg{
		Ctx:             ctx,
		ResponseChannel: respChannel,
		SvcInfo:         svcInfo,
		Op:              servertypes.Unregister,
	}
	// send it to the worker via our queue
	registrationQueue.Add(regMsg)
	// wait for the response
	respMsg := <-respChannel
	// we expect the worker to close the channel after the response has been sent
	return respMsg.Result, respMsg.Err
}

// function to create a copy of the ServiceInfo
func copy(sInfo *reg.ServiceInfo) *reg.ServiceInfo {
	c := &reg.ServiceInfo{
		Namespace:   sInfo.Namespace,
		ServiceName: sInfo.ServiceName,
		HostName:    sInfo.HostName,
		Ipaddress:   sInfo.Ipaddress,
		NodeName:    sInfo.NodeName,
		Ports:       sInfo.Ports,
		Weight:      sInfo.Weight,
	}
	return c
}

func main() {
	// TODO: implement a better logging solution
	// initialize our Kubernetes client
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	clientset, err := kclient.InitKubeClient(kubeconfig)
	if err != nil {
		klog.Fatalf("FATAL: cannot initialize Kubernetes client: %s", err.Error())
	}

	klog.Info("Creating IndexInformer for endpoints objects in all namespaces")
	indexInformer := epwatcher.CreateIndexInformer(clientset)
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

	registrationWorker := worker.NewWorker(registrationQueue, (*indexInformer).GetIndexer())
	go (*registrationWorker).Run(stop)

	if !syncError {
		klog.Infof("listening for requests on localhost%s ...\n", port)
		lis, err := net.Listen("tcp", port)
		if err != nil {
			klog.Fatalf("failed to listen : %v", err)
		}
		s := grpc.NewServer()
		reg.RegisterServiceRegistryService(s, &reg.ServiceRegistryService{Register: register, UnRegister: unRegister})
		if err := s.Serve(lis); err != nil {
			klog.Fatalf("failed to serve: %v", err)
		}
	} else {
		klog.Fatal("Cache sync error unrecoverable - exiting!")
	}
}
