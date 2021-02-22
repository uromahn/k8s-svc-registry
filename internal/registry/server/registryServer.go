package registry

import (
	"context"

	"k8s.io/klog/v2"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"

	kclient "github.com/uromahn/k8s-svc-registry/internal/kubeclient"
	servertypes "github.com/uromahn/k8s-svc-registry/internal/servertypes"
	"k8s.io/client-go/util/workqueue"
)

type ServiceRegistryServer struct {
	reg.UnimplementedServiceRegistryServer
}

var registrationQueue workqueue.RateLimitingInterface

func InitRegistryServer(queue workqueue.RateLimitingInterface) {
	registrationQueue = queue
}

// register implements registry.ServiceRegistryService.Register
func (*ServiceRegistryServer) Register(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
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
func (*ServiceRegistryServer) UnRegister(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
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
