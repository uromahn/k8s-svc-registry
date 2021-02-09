package endpointswatcher

import (
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func CreateIndexInformer(clientset *kubernetes.Clientset) *cache.SharedIndexInformer {
	klog.Info("Starting endpointswatcher")
	labelSelector := "service-type=external"
	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = labelSelector
	}

	watchlist := cache.NewFilteredListWatchFromClient(
		clientset.CoreV1().RESTClient(),
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
				klog.Infof("endpoints added: %s \n", obj)
				key := "test-dev/test"
				item, exists, _ := sharedIndexInformer.GetIndexer().GetByKey(key)
				if exists {
					klog.Infof("found key %s with item %v in cache", key, item)
				} else {
					klog.Infof("No item with key %s found in cache", key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				klog.Infof("endpoints deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				klog.Infof("endpoints changed:\n %s ->\n %s\n", oldObj, newObj)
				key := "test-dev/test"
				item, exists, _ := sharedIndexInformer.GetIndexer().GetByKey(key)
				if exists {
					klog.Infof("found key %s with item %v in cache", key, item)
				} else {
					klog.Infof("No item with key %s found in cache", key)
				}
			},
		},
	)
	klog.Infof("Endpoints list watcher and index informer created...")
	return &sharedIndexInformer
}
