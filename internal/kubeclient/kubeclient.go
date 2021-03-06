package kubeclient

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"
)

type KubeClient struct {
	Clientset *kubernetes.Clientset
}

// InitKubeClient initializes a Kubernetes clientset with the given kubeconfig
func NewKubeClient(kubeconfig *string, clientset *kubernetes.Clientset) (*KubeClient, error) {
	var cs *kubernetes.Clientset
	if clientset == nil {
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			klog.Errorf("ERROR - while creating the config: %s", err.Error())
			return nil, err
		}
		cs, err = kubernetes.NewForConfig(config)
		if err != nil {
			klog.Errorf("ERROR - while creating clientset: %s", err.Error())
			return nil, err
		}
	} else {
		cs = clientset
	}

	return &KubeClient{Clientset: cs}, nil
}

// GetOrCreateNamespace will try to get the namespace object named 'nsName', or
// if it doesn't exist, will create it.
func (kc *KubeClient) GetOrCreateNamespace(ctx context.Context, nsName string, errorIfNotExist bool) (*apiv1.Namespace, error) {
	// first try to get the namespace to see if it already exists
	nsClient := kc.Clientset.CoreV1().Namespaces()
	ns, err := nsClient.Get(ctx, nsName, metav1.GetOptions{})
	if err == nil {
		klog.Infof("INFO: namespace '%s' exists.", nsName)
		// we have the namespace already, just return it
		return ns, err
	}
	if k8serrors.IsNotFound(err) {
		if errorIfNotExist {
			errMsg := fmt.Sprintf("namepace %s does not exist", nsName)
			klog.Error(errMsg)
			return nil, errors.New(errMsg)
		}
		// otherwise we need to create it
		klog.Infof("INFO: creating namepsace '%s'", nsName)
		nsMeta := createNamespaceSpec(nsName)
		ns, err = nsClient.Create(ctx, nsMeta, metav1.CreateOptions{})
		return ns, err
	}
	return nil, err
}

// GetOrCreateService will try to get a service with the name 'svcName'
// in the namespace 'ns'. If it doesn't exist, it will create the object.
func (kc *KubeClient) GetOrCreateService(ctx context.Context, ns string, svcName string, ports []*reg.NamedPort, errorIfNotExist bool) (*apiv1.Service, error) {
	svcClient := kc.Clientset.CoreV1().Services(ns)
	svc, err := svcClient.Get(ctx, svcName, metav1.GetOptions{})
	if err == nil {
		klog.Infof("INFO: service '%s' in namespace '%s' exists.", svcName, ns)
		return svc, err
	}
	if k8serrors.IsNotFound(err) {
		if errorIfNotExist {
			errMsg := fmt.Sprintf("service %s in namespace %s does not exist", svcName, ns)
			klog.Error(errMsg)
			return nil, errors.New(errMsg)
		}
		klog.Infof("INFO: creating service '%s' in namespace '%s' with named ports '%v'", svcName, ns, ports)
		svcSpec := createServiceSpec(svcName, ns, ports)
		svc, err = svcClient.Create(ctx, svcSpec, metav1.CreateOptions{})
		return svc, err
	}
	return nil, err
}

// CreateNewEndpoint is a function that creates a new endpoints object
// and adds the service given in svcInfo to it.
// This function makes the following assumptions:
// * the namespace given in nsName already exists
// * a headless service as provided in svcInfo.ServiceName has already been created
func (kc *KubeClient) CreateNewEndpoint(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	// before we do anything, we check if the service object actually exists
	exists, err := kc.checkSvcExists(ctx, svcInfo)
	if !exists {
		errMsg := fmt.Sprint("service does not exist")
		result := newRegistrationResult(svcInfo, uint32(404), errMsg)
		return result, err
	}
	epClient := kc.Clientset.CoreV1().Endpoints(svcInfo.Namespace)
	newEp := newEndpointsObj(ctx, svcInfo)
	statusCode := uint32(200)
	statusDetails := "registered"
	_, err = epClient.Create(ctx, &newEp, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("unable to create new Endpoints object - %s", err.Error())
		statusCode = uint32(500)
		statusDetails = err.Error()
	}
	result := newRegistrationResult(svcInfo, statusCode, statusDetails)
	return result, err
}

// AddSvcToEndpoint is a function that adds the service given in svcInfo
// in the corresponding endpoints object of that service.
// This function makes the following assumptions:
// * the namespace given in nsName already exists
// * a headless service as provided in svcInfo.ServiceName has already been created
// * an existing non-nil endpoints object ep is being passed in
//
// We have to consider the following cases:
// 1. The endpoints object already exists and
//    a. the Ipaddress:Ports already exists: we do nothing since the svcInfo has already been registered
//    b. the Ipaddress:ports already exists in the notReadyAddresses array: move the EndpointAddress to the addresses array
// 2. The endpoints object already exists and the service endpoint does not exist yet.
//    There are two subcases to be considered:
//    a. the given svcInfo.Ports already exists: we then have to add the new svcInfo to the addresses array
//    b. the given svcInfo.Ports does not exist: we have to create a new endpointsSubset object with the Ipaddress:Ports
//
// TODO: in a future refactoring, we might want to explore if we can benefit from this library: https://github.com/banzaicloud/k8s-objectmatcher
func (kc *KubeClient) AddSvcToEndpoint(ctx context.Context, epClient v1.EndpointsInterface, ep *apiv1.Endpoints, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	var err error = nil
	// let's add our service info to the endpoints object
	newIP := svcInfo.Ipaddress
	newPorts := svcInfo.Ports
	subsets := ep.Subsets

	if ep == nil {
		errMsg := fmt.Sprint("passed endpoints object nil")
		klog.Error(errMsg)
		result := newRegistrationResult(svcInfo, uint32(500), errMsg)
		return result, err
	}
	// before we do anything, we check if the service object actually exists
	exists, err := kc.checkSvcExists(ctx, svcInfo)
	if !exists {
		errMsg := fmt.Sprint("service does not exist")
		result := newRegistrationResult(svcInfo, uint32(404), errMsg)
		return result, err
	}
	if epClient == nil {
		epClient = kc.Clientset.CoreV1().Endpoints(svcInfo.Namespace)
	}
	// create a map from our ports array to check for equality
	portsMap := make(map[int32]bool, len(newPorts))
	for _, p := range newPorts {
		portsMap[p.GetPort()] = true
	}
	l := 1
	if len(subsets) > 0 {
		l = len(subsets)
	}
	addrsSubsets := make([][]apiv1.EndpointAddress, l, l)
	nrAddrsSubsets := make([][]apiv1.EndpointAddress, l, l)
	portsSubsets := make([][]apiv1.EndpointPort, l, l)

	// extract the subsets which contain a list of EndpointSubset objects
	for i, endpointSubset := range subsets {
		addrsSubsets[i] = endpointSubset.Addresses
		nrAddrsSubsets[i] = endpointSubset.NotReadyAddresses
		portsSubsets[i] = endpointSubset.Ports
	}
	// check if there is a subset with our ports already
	subsetWithPort := -1
	for i, portsSubset := range portsSubsets {
		if samePorts(portsSubset, portsMap) {
			subsetWithPort = i
			// break out of the loop
			break
		}
	}
	var result *reg.RegistrationResult = nil
	err = nil
	if subsetWithPort != -1 {
		// we have the ports already, so check if we have the IP address already in the not ready address list
		notReadyIPPos := -1
		for pos, addr := range nrAddrsSubsets[subsetWithPort] {
			if addr.IP == newIP {
				notReadyIPPos = pos
				break
			}
		}
		// Case #1.b.
		if notReadyIPPos != -1 {
			// move the endpoint from the notReady address subsets to the ready addressSubset
			moveEndpointAddress(notReadyIPPos, &nrAddrsSubsets[subsetWithPort], &addrsSubsets[subsetWithPort])
			_, err = epClient.Update(ctx, ep, metav1.UpdateOptions{})
			statusCode := uint32(200)
			statusDetails := "registered"
			if err != nil {
				klog.Errorf("unable to move Endpoints object from notReady to ready - %s", err.Error())
				statusCode = 500
				statusDetails = "unable to register"
			}
			result = newRegistrationResult(svcInfo, statusCode, statusDetails)
			return result, err
		}
		// check if the IP address already exists in the ready list
		alreadyExists := false
		for _, addr := range addrsSubsets[subsetWithPort] {
			if addr.IP == newIP {
				alreadyExists = true
				break
			}
		}
		// case #1.a.
		if alreadyExists {
			// the endpoint already exists - do nothing and just return the result object
			result = newRegistrationResult(svcInfo, uint32(200), "registered")
		} else {
			// Case #3.a.
			newEpAddress := apiv1.EndpointAddress{
				IP:        svcInfo.Ipaddress,
				Hostname:  svcInfo.HostName,
				NodeName:  &svcInfo.NodeName,
				TargetRef: nil,
			}
			ep.Subsets[subsetWithPort].Addresses = append(ep.Subsets[subsetWithPort].Addresses, newEpAddress)
			_, err = epClient.Update(ctx, ep, metav1.UpdateOptions{})
			statusCode := uint32(200)
			statusDetails := "registered"
			if err != nil {
				klog.Errorf("unable to append Endpoints object - %s", err.Error())
				statusCode = 500
				statusDetails = "unable to register"
			}
			result = newRegistrationResult(svcInfo, statusCode, statusDetails)
		}
	} else {
		// Case #2.b.
		// we need to create a new EndpointSubset with the port and the EndpointAddress here
		newSubset := newEndpointsSubsetObj(ctx, svcInfo)
		ep.Subsets = append(ep.Subsets, newSubset)
		_, err = epClient.Update(ctx, ep, metav1.UpdateOptions{})
		statusCode := uint32(200)
		statusDetails := "registered"
		if err != nil {
			klog.Errorf("unable to create and add new EndpointsSubset object - %s", err.Error())
			statusCode = 500
			statusDetails = "unable to register"
		}
		result = newRegistrationResult(svcInfo, statusCode, statusDetails)
	}
	return result, err
}

// RegisterEndpoint is a function that will ensure that the service given in svcInfo
// will be registered in the corresponding endpoints object of that service.
// This function makes the following assumptions:
// * the namespace given in nsName already exists
// * a headless service as provided in svcInfo.ServiceName has already been created
//
// We have to consider the following cases:
// 1. This is the first service endpoint to be registered:
//    the endpoints object does not exist so we create the endpoints object with the svcInfo added
// 2. The endpoints object already exists and
//    a. the Ipaddress:Ports already exists: we do nothing since the svcInfo has already been registered
//    b. the Ipaddress:ports already exists in the notReadyAddresses array: move the EndpointAddress to the addresses array
// 3. The endpoints object already exists and the service endpoint does not exist yet.
//    There are two subcases to be considered:
//    a. the given svcInfo.Ports already exists: we then have to add the new svcInfo to the addresses array
//    b. the given svcInfo.Ports does not exist: we have to create a new endpointsSubset object with the Ipaddress:Ports
//
// TODO: in a future refactoring, we might want to explore if we can benefit from this library: https://github.com/banzaicloud/k8s-objectmatcher
func (kc *KubeClient) RegisterEndpoint(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	// before we do anything, we check if the service object actually exists
	exists, err := kc.checkSvcExists(ctx, svcInfo)
	if !exists {
		errMsg := fmt.Sprint("service does not exist")
		result := newRegistrationResult(svcInfo, uint32(404), errMsg)
		return result, err
	}
	epClient := kc.Clientset.CoreV1().Endpoints(svcInfo.Namespace)
	// let's see if we already have an Endpoints object for our service
	ep, err := epClient.Get(ctx, svcInfo.ServiceName, metav1.GetOptions{})

	// Case #1: no endpoint exists for the service yet
	if k8serrors.IsNotFound(err) {
		// we don't have the endpoints object already, so we need to create it here
		newEp := newEndpointsObj(ctx, svcInfo)
		statusCode := uint32(200)
		statusDetails := "registered"
		_, err := epClient.Create(ctx, &newEp, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("unable to create new Endpoints object - %s", err.Error())
			statusCode = uint32(500)
			statusDetails = err.Error()
		}
		result := newRegistrationResult(svcInfo, statusCode, statusDetails)
		return result, err
	}

	// Case #2: we have the current endpoints object for the service now
	if err == nil {
		return kc.AddSvcToEndpoint(ctx, epClient, ep, svcInfo)
	}
	// if we land here, some error happened when trying to get the endpoints object
	return nil, err
}

// UnregisterWithEndpoint is a function that will remove the service given in svcInfo
// from the passed in endpoints object ep of that service.
// This function makes the following assumptions:
// * the namespace given in nsName already exists
// * a headless service as provided in svcInfo.ServiceName exists
// * an existing non-nil endpoints object ep is being passed in
//
// TODO: in a future refactoring, we might want to explore if we can benefit from this library: https://github.com/banzaicloud/k8s-objectmatcher
func (kc *KubeClient) UnregisterWithEndpoint(ctx context.Context, epClient v1.EndpointsInterface, ep *apiv1.Endpoints, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	var result *reg.RegistrationResult = nil
	var err error = nil
	svcIP := svcInfo.Ipaddress
	svcPorts := svcInfo.Ports

	if ep == nil {
		errMsg := fmt.Sprint("passed endpoints object nil")
		klog.Error(errMsg)
		result := newRegistrationResult(svcInfo, uint32(500), errMsg)
		return result, err
	}
	exists, err := kc.checkSvcExists(ctx, svcInfo)
	if !exists {
		errMsg := fmt.Sprint("service does not exist")
		result := newRegistrationResult(svcInfo, uint32(404), errMsg)
		return result, err
	}

	if epClient == nil {
		epClient = kc.Clientset.CoreV1().Endpoints(svcInfo.Namespace)
	}

	subsets := ep.Subsets

	// create a map from our ports array to check for equality
	portsMap := make(map[int32]bool, len(svcPorts))
	for _, p := range svcPorts {
		portsMap[p.GetPort()] = true
	}
	l := 1
	if len(subsets) > 0 {
		l = len(subsets)
	}
	addrsSubsets := make([][]apiv1.EndpointAddress, l, l)
	nrAddrsSubsets := make([][]apiv1.EndpointAddress, l, l)
	portsSubsets := make([][]apiv1.EndpointPort, l, l)

	// extract the subsets which contain a list of EndpointSubset objects
	for i, endpointSubset := range subsets {
		addrsSubsets[i] = endpointSubset.Addresses
		nrAddrsSubsets[i] = endpointSubset.NotReadyAddresses
		portsSubsets[i] = endpointSubset.Ports
	}
	// check if a subset with our ports exist
	subsetWithPort := -1
	for i, portsSubset := range portsSubsets {
		if samePorts(portsSubset, portsMap) {
			subsetWithPort = i
			// break out of the loop
		}
	}
	if subsetWithPort != -1 {
		// we found the ports, so check if we have the IP address in the address list
		readyIPPos := -1
		for pos, addr := range addrsSubsets[subsetWithPort] {
			if addr.IP == svcIP {
				readyIPPos = pos
				break
			}
		}
		if readyIPPos != -1 {
			// delete the endpointAddress from the subset
			_ = deleteEndpointAddress(readyIPPos, &ep.Subsets[subsetWithPort].Addresses)
			if len(ep.Subsets[subsetWithPort].Addresses) == 0 && len(ep.Subsets[subsetWithPort].NotReadyAddresses) == 0 {
				ep.Subsets[subsetWithPort] = ep.Subsets[len(ep.Subsets)-1]
				ep.Subsets[len(ep.Subsets)-1] = apiv1.EndpointSubset{}
				ep.Subsets = ep.Subsets[:len(ep.Subsets)-1]
			}
			if len(ep.Subsets) == 0 {
				// we have no subset left, so delete the entire endpoints object
				err = epClient.Delete(ctx, ep.Name, metav1.DeleteOptions{})
			} else {
				_, err = epClient.Update(ctx, ep, metav1.UpdateOptions{})
			}
			statusCode := uint32(204)
			statusDetails := "unregistered"
			if err != nil {
				klog.Errorf("unable to remove addresses from endpoints object - %s", err.Error())
				statusCode = uint32(500)
				statusDetails = "unable to unregister"
			}
			result = newRegistrationResult(svcInfo, statusCode, statusDetails)
			return result, err
		}
	}
	// we don't seem to have a service with those ports registered, so respond accordingly
	statusCode := uint32(404)
	statusDetails := "service endpoint not found"
	result = newRegistrationResult(svcInfo, statusCode, statusDetails)
	return result, nil
}

// UnregisterEndpoint is a function that will ensure that the service given in svcInfo
// will be unregistered from the corresponding endpoints object of that service.
// This function makes the following assumptions:
// * the namespace given in nsName already exists
// * a headless service as provided in svcInfo.ServiceName exists
//
// TODO: in a future refactoring, we might want to explore if we can benefit from this library: https://github.com/banzaicloud/k8s-objectmatcher
func (kc *KubeClient) UnregisterEndpoint(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	var result *reg.RegistrationResult = nil
	var err error = nil

	exists, err := kc.checkSvcExists(ctx, svcInfo)
	if !exists {
		errMsg := fmt.Sprint("service does not exist")
		result := newRegistrationResult(svcInfo, uint32(404), errMsg)
		return result, err
	}
	epClient := kc.Clientset.CoreV1().Endpoints(svcInfo.Namespace)
	// let's retrieve the Endpoints object for our service
	ep, err := epClient.Get(ctx, svcInfo.ServiceName, metav1.GetOptions{})

	// the endpoints object does not exist: respond with a "not found"
	if k8serrors.IsNotFound(err) {
		klog.Errorf("Endpoints objectfor service %s in namespace %s does not exist", svcInfo.ServiceName, svcInfo.Namespace)
		statusCode := uint32(404)
		statusDetails := "service endpoint not found"
		result = newRegistrationResult(svcInfo, statusCode, statusDetails)
		return result, nil
	}

	if err == nil {
		return kc.UnregisterWithEndpoint(ctx, epClient, ep, svcInfo)
	}
	return nil, err
}

func (kc *KubeClient) checkSvcExists(ctx context.Context, svcInfo *reg.ServiceInfo) (bool, error) {
	var exists bool = false
	svcClient := kc.Clientset.CoreV1().Services(svcInfo.Namespace)
	_, err := svcClient.Get(ctx, svcInfo.ServiceName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		klog.Infof("%s: service %s in namespace %s does not exists! Cannot register endpoint for IP %s ports %v",
			err.Error(), svcInfo.ServiceName, svcInfo.Namespace, svcInfo.Ipaddress, svcInfo.Ports)
		exists = false
	} else {
		exists = true
	}
	return exists, err
}

func createNamespaceSpec(ns string) *apiv1.Namespace {
	nsSpec := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	return nsSpec
}

func createServiceSpec(svcName string, ns string, ports []*reg.NamedPort) *apiv1.Service {
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

// samePorts returns true if the numerical part of the ports is the same.
// The arrays aren't necessarily sorted so we (re)use a map.
func samePorts(ep []apiv1.EndpointPort, portsMap map[int32]bool) bool {
	if len(ep) != len(portsMap) {
		return false
	}
	for _, e := range ep {
		if !portsMap[e.Port] {
			return false
		}
	}
	return true
}

func newRegistrationResult(svcInfo *reg.ServiceInfo, statusCode uint32, statusDetails string) *reg.RegistrationResult {
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

func moveEndpointAddress(fromIndex int, fromAddr *[]apiv1.EndpointAddress, toAddr *[]apiv1.EndpointAddress) {
	fromEpAddress := (*fromAddr)[fromIndex]
	// append the EndpointAddress from the "from" slice to the "to" slice
	toAppended := append(*toAddr, fromEpAddress)
	*toAddr = toAppended
	// now delete the endpointAddress at index "fromIndex" from the "fromAddr" slice
	fromAddr = deleteEndpointAddress(fromIndex, fromAddr)
}

// This method deletes the EndpointAddress at index 'index' from the 'addrArray'
// IMPORTANT: this does not preserve the original order of the array
func deleteEndpointAddress(index int, addrArray *[]apiv1.EndpointAddress) *[]apiv1.EndpointAddress {
	// remove the element at index by overwriting it with the last element
	(*addrArray)[index] = (*addrArray)[len(*addrArray)-1]
	// overwrite the last element with an empty element to avoid a memory leak
	(*addrArray)[len(*addrArray)-1] = apiv1.EndpointAddress{}
	// return the slice minus the last element
	*addrArray = (*addrArray)[:len(*addrArray)-1]
	return addrArray
}

func newEndpointsObj(ctx context.Context, svcInfo *reg.ServiceInfo) apiv1.Endpoints {
	endpoints := apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcInfo.ServiceName,
			Namespace: svcInfo.Namespace,
			Labels: map[string]string{
				"service-type": "external",
			},
		},
		Subsets: []apiv1.EndpointSubset{
			newEndpointsSubsetObj(ctx, svcInfo),
		},
	}
	return endpoints
}

func newEndpointsSubsetObj(ctx context.Context, svcInfo *reg.ServiceInfo) apiv1.EndpointSubset {
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
