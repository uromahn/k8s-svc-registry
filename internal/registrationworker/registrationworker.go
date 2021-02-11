package registrationworker

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kclient "github.com/uromahn/k8s-svc-registry/internal/kubeclient"
	servertypes "github.com/uromahn/k8s-svc-registry/internal/servertypes"
)

type Worker struct {
	queue workqueue.RateLimitingInterface
	cache cache.Indexer
}

func NewWorker(queue workqueue.RateLimitingInterface, cache cache.Indexer) *Worker {
	return &Worker{
		queue: queue,
		cache: cache,
	}
}

func (w *Worker) Run(stopCh chan struct{}) {
	defer runtime.HandleCrash()
	defer w.queue.ShutDown()

	klog.Info("Starting registration worker")

	go wait.Until(w.execWorker, time.Second, stopCh)
	<-stopCh

	klog.Info("Stopping registration worker")
}

func (w *Worker) execWorker() {
	for w.processNextRegistration() {
	}
}

func (w *Worker) processNextRegistration() bool {
	// Block and wait until there is a new item in the working queue
	obj, quit := w.queue.Get()

	if quit {
		return false
	}

	defer w.queue.Done(obj)

	// convert our object 'obj' to the expected struct
	msg, ok := obj.(servertypes.RegistrationMsg)
	if ok {
		err := w.doWork(msg)
		// handle an error in case the registration failed for whatever reason
		w.handleError(err, msg)
	}
	return true
}

func (w *Worker) doWork(msg servertypes.RegistrationMsg) error {
	var err error = nil
	svcInfo := msg.SvcInfo
	respChan := msg.ResponseChannel
	ns := svcInfo.Namespace
	svc := svcInfo.ServiceName
	key := ns + "/" + svc
	op := msg.Op
	ctx := msg.Ctx

	obj, exists, err := w.cache.GetByKey(key)

	if err == nil {
		if op == servertypes.Register {
			if exists {
				ep, ok := obj.(*apiv1.Endpoints)
				if !ok {
					errMsg := fmt.Sprintf("Cached object is of type %T which does not match expected type of v1.Endpoints", obj)
					klog.Error(errMsg)
					err = fmt.Errorf(errMsg)
					resultMsg := servertypes.ResultMsg{
						Result: nil,
						Err:    err,
					}
					respChan <- resultMsg
				} else {
					result, err := kclient.AddSvcToEndpoint(ctx, nil, ep, svcInfo)
					w.handleError(err, msg)
					if err == nil {
						resultMsg := servertypes.ResultMsg{
							Result: result,
							Err:    err,
						}
						respChan <- resultMsg
					}
				}
			} else {
				result, err := kclient.CreateNewEndpoint(ctx, svcInfo)
				w.handleError(err, msg)
				if err == nil {
					resultMsg := servertypes.ResultMsg{
						Result: result,
						Err:    err,
					}
					respChan <- resultMsg
				}
			}
		} else if op == servertypes.Unregister {
			if exists {
				ep, ok := obj.(*apiv1.Endpoints)
				if !ok {
					errMsg := fmt.Sprintf("Cached object is of type %T which does not match expected type of v1.Endpoints", obj)
					klog.Error(errMsg)
					err = fmt.Errorf(errMsg)
					resultMsg := servertypes.ResultMsg{
						Result: nil,
						Err:    err,
					}
					respChan <- resultMsg
				} else {
					result, err := kclient.UnregisterWithEndpoint(ctx, nil, ep, svcInfo)
					w.handleError(err, msg)
					if err == nil {
						resultMsg := servertypes.ResultMsg{
							Result: result,
							Err:    err,
						}
						respChan <- resultMsg
					}
				}
			} else {
				errMsg := fmt.Sprintf("Can't unregister endpoint %s:%s for %s - endpoints object does not exist", msg.SvcInfo.HostName, msg.SvcInfo.Ipaddress, key)
				klog.Warning(errMsg)
				err = fmt.Errorf(errMsg)
				resultMsg := servertypes.ResultMsg{
					Result: nil,
					Err:    err,
				}
				respChan <- resultMsg
			}
		} else {
			// we have an unknown operation, error out here
			errMsg := fmt.Sprintf("Unknown operation %d received for svcInfo %v", op, msg.SvcInfo)
			klog.Error(errMsg)
			err = fmt.Errorf(errMsg)
			resultMsg := servertypes.ResultMsg{
				Result: nil,
				Err:    err,
			}
			respChan <- resultMsg
		}
	} else {
		errMsg := fmt.Sprintf("Fetching endpoints '%s' from cache failed with error %s", key, err.Error())
		klog.Error(errMsg)
		err = fmt.Errorf(errMsg)
		resultMsg := servertypes.ResultMsg{
			Result: nil,
			Err:    err,
		}
		respChan <- resultMsg
	}
	return err
}

func (w *Worker) handleError(err error, msg servertypes.RegistrationMsg) {
	if err == nil {
		// Forget about the #AddRateLimited history of the msg on every successful synchronization.
		// This ensures that future processing of this msg is not delayed because of
		// an outdated error history.
		w.queue.Forget(msg)
		return
	}

	// This worker retries 5 times if something goes wrong. After that, it stops trying.
	if w.queue.NumRequeues(msg) < 5 {
		klog.Infof("Error processing registration requuest %v: %v", msg, err)

		// Re-enqueue the message rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the message will be processed later again.
		w.queue.AddRateLimited(msg)
		return
	}

	w.queue.Forget(msg)
	// Report to an external entity that, even after several retries, we could not successfully process this message
	runtime.HandleError(err)
	klog.Infof("Dropping registration request %v out of the queue: %v", msg, err)
	resultMsg := servertypes.ResultMsg{
		Result: nil,
		Err:    err,
	}
	msg.ResponseChannel <- resultMsg
}
