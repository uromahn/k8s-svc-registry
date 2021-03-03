package registry

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"k8s.io/klog/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"
	epwatcher "github.com/uromahn/k8s-svc-registry/internal/endpointswatcher"
	kclient "github.com/uromahn/k8s-svc-registry/internal/kubeclient"
	worker "github.com/uromahn/k8s-svc-registry/internal/registrationworker"
	registryC "github.com/uromahn/k8s-svc-registry/internal/registry/client"
	registryS "github.com/uromahn/k8s-svc-registry/internal/registry/server"
)

// here come a bunch of helper functions
// ---------------------------------------------------------------------------------
func namespace(name string) *apiv1.Namespace {
	nsSpec := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	return nsSpec
}

func service(svcName string, ns string, ports []*reg.NamedPort) *apiv1.Service {
	var specPorts []apiv1.ServicePort

	// populate our spectPorts array
	for _, namedPort := range ports {
		port := apiv1.ServicePort{
			Name:     namedPort.GetName(),
			Protocol: apiv1.ProtocolTCP,
			Port:     namedPort.GetPort(),
		}
		specPorts = append(specPorts, port)
	}
	svcSpec := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
			Labels: map[string]string{
				"service-type": "external",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports:     specPorts,
			ClusterIP: "None",
			Type:      apiv1.ServiceTypeClusterIP,
		},
	}
	return svcSpec
}
func endpoints(svcInfo *reg.ServiceInfo) apiv1.Endpoints {
	endpoints := apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcInfo.ServiceName,
			Namespace: svcInfo.Namespace,
			Labels: map[string]string{
				"service-type": "external",
			},
		},
		Subsets: []apiv1.EndpointSubset{
			endpointsSubset(svcInfo),
		},
	}
	return endpoints
}

func endpointsSubset(svcInfo *reg.ServiceInfo) apiv1.EndpointSubset {
	var subsetPorts []apiv1.EndpointPort
	// populate the subsetPorts array
	for _, namedPort := range svcInfo.GetPorts() {
		port := apiv1.EndpointPort{
			Name:     namedPort.GetName(),
			Port:     namedPort.GetPort(),
			Protocol: apiv1.ProtocolTCP,
		}
		subsetPorts = append(subsetPorts, port)
	}
	subset := apiv1.EndpointSubset{
		Addresses: []apiv1.EndpointAddress{
			{
				IP:       svcInfo.Ipaddress,
				Hostname: svcInfo.HostName,
				NodeName: &svcInfo.NodeName,
			},
		},
		NotReadyAddresses: []apiv1.EndpointAddress{},
		Ports:             subsetPorts,
	}
	return subset
}

func registrationResult(svcInfo *reg.ServiceInfo, statusCode uint32, statusDetails string) *reg.RegistrationResult {
	result := &reg.RegistrationResult{
		Namespace:     svcInfo.GetNamespace(),
		ServiceName:   svcInfo.GetServiceName(),
		Ipaddress:     svcInfo.GetIpaddress(),
		Ports:         svcInfo.GetPorts(),
		Status:        statusCode,
		StatusDetails: statusDetails,
	}
	return result
}

// end of helper functions
// ---------------------------------------------------------------------------------

// some test data
const (
	nsName  string = "test-dev"
	svcName string = "test"
)

var port = &reg.NamedPort{
	Port: 9080,
	Name: "http",
}

var namedPorts = []*reg.NamedPort{port}

var testNs = namespace(nsName)
var testSvc = service(svcName, nsName, namedPorts)

var svcInfo = &reg.ServiceInfo{
	Namespace:   nsName,
	ServiceName: svcName,
	HostName:    "localhost",
	Ipaddress:   "10.0.0.1",
	NodeName:    "uromahn-vm-ubuntu18",
	Ports:       namedPorts,
	Weight:      1.0,
}

var tests = []struct {
	description string
	namespace   string
	request     *reg.ServiceInfo
	expected    *reg.RegistrationResult
	objs        []runtime.Object
}{
	{"no ns - no service", nsName, svcInfo, nil, nil},
	{"with ns - no service", nsName, svcInfo, nil, []runtime.Object{testNs}},
	{"with ns - with service", nsName, svcInfo, nil, []runtime.Object{testNs, testSvc}},
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	reg.RegisterServiceRegistryServer(server, &registryS.ServiceRegistryServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			klog.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		klog.Fatal(err)
	}
	defer conn.Close()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(test.objs...)
			k8sClient, err := kclient.NewKubeClient(nil, clientset.CoreV1())
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
				utilRuntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
				syncError = true
			}

			klog.Info("Creating workqueue to process new service registrations")
			registrationQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
			registryS.InitRegistryServer(registrationQueue, k8sClient)

			registrationWorker := worker.NewWorker(registrationQueue, (*indexInformer).GetIndexer(), k8sClient)
			go (*registrationWorker).Run(stop)

			if !syncError {
				grpcClient := registryC.NewServiceRegistryClient(conn, time.Duration(10)*time.Second)
				actualResult, err := grpcClient.Register(ctx, test.request)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				t.Logf("actual result = %v", actualResult)
			} else {
				klog.Fatal("Cache sync error unrecoverable - exiting!")
			}
		})
	}
}
