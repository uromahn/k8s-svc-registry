package endpointswatcher

/*
The following code has been "borrowed" from the client-go workqueue example
since it can be used here almost unchanged
*/

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller is used to watch endpoints and puts them into an indexed cache
type Controller struct {
	Indexer  cache.Indexer
	Queue    workqueue.RateLimitingInterface
	Informer cache.Controller
	StopChan chan struct{}
}

// NewController creates a new Controller.
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, stop chan struct{}) *Controller {
	return &Controller{
		Informer: informer,
		Indexer:  indexer,
		Queue:    queue,
		StopChan: stop,
	}
}

// TODO: modify this function since we are not really processing anything here
//
func (c *Controller) processNextItem() bool {
	klog.Infof("processNextItem")
	// Wait until there is a new item in the working queue
	key, quit := c.Queue.Get()
	klog.Infof("received new item to process...")
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.Queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// TODO: we may want to modify this here or even remove this function entirely
//
// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.Indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Endpoints %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		fmt.Printf("Sync/Add/Update for Endpoints %s\n", obj.(*apiv1.Endpoints).GetName())
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.Queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.Queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing endpoints %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.Queue.AddRateLimited(key)
		return
	}

	c.Queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping endpoints %q out of the queue: %v", key, err)
}

// Run begins watching and syncing.
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.Queue.ShutDown()
	klog.Info("Starting Endpoints controller")

	go c.Informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.Informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	klog.Infof("caches completed initial sync")
	// we may want to change this since we don't really do anything here!
	//
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping Endpoints controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

// Endpointswatcher a watcher that monitors endpoints in all namespaces inour K8s server and caches
// their state inside an indexer.
func Endpointswatcher(clientset *kubernetes.Clientset) (*Controller, error) {
	klog.Infof("Setting up Endpointswatcher.")
	// create the endpoints watcher
	klog.Infof("setting up listwatcher for endpoints in all namespaces...")
	endpointsListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "endpoints", "test-dev", fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	klog.Infof("creating new SharedIndexInformer...")
	indexInformer := cache.NewSharedIndexInformer(endpointsListWatcher, &apiv1.Endpoints{}, time.Second, cache.Indexers{})
	rehf := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}
	klog.Infof("adding ResourceEventHandlerFuncs...")
	indexInformer.AddEventHandler(rehf)
	klog.Infof("creating stop channel...")
	stop := make(chan struct{})
	klog.Infof("creating new controller...")
	controller := NewController(queue, indexInformer.GetIndexer(), indexInformer.GetController(), stop)

	klog.Infof("starting controller!")
	go controller.Run(1, stop)
	klog.Infof("Setup of Endpointswatcher completed.")

	return controller, nil
}
