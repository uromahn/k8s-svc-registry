package endpointswatcher

import (
	"k8s.io/klog/v2"

	kclient "github.com/uromahn/k8s-svc-registry/internal/kubeclient"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// CreateIndexInformer function to create a new IndexInformer
func CreateIndexInformer(k8sClient *kclient.KubeClient) *cache.SharedIndexInformer {
	klog.Info("Starting endpointswatcher")
	labelSelector := "service-type=external"
	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = labelSelector
	}

	watchlist := cache.NewFilteredListWatchFromClient(
		k8sClient.Clientset.CoreV1().RESTClient(),
		"endpoints",
		apiv1.NamespaceAll,
		optionsModifier,
	)
	sharedIndexInformer := cache.NewSharedIndexInformer(
		watchlist,
		&apiv1.Endpoints{},
		0,
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
	)
	sharedIndexInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				klog.Info("endpoints added")
			},
			DeleteFunc: func(obj interface{}) {
				klog.Info("endpoints deleted")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				klog.Info("endpoints changed")
			},
		},
	)
	klog.Infof("Endpoints list watcher and index informer created...")
	return &sharedIndexInformer
}
