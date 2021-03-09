package util

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// NewFilteredEndpointsListWatchFromClient creates a new ListWatch from the specified client, resource, namespace, and option modifier.
// Option modifier is a function that takes a ListOptions and modifies the consumed ListOptions. Provide customized modifier function
// to apply modification to ListOptions with a field selector, a label selector, or any other desired options.
//
// This function was copied over from the NewFilteredListWatchFromClient in https://github.com/kubernetes/client-go/blob/master/tools/cache/listwatch.go
// and modified to use a v1.CoreV1Interface instead of a RESTClient. This required us to make the function specific for the Endoints type
// instead of having it generic for all types. However, this function also works with the Fake client.
func NewFilteredEndpointsListWatchFromClient(ctx context.Context, c v1.CoreV1Interface, namespace string, optionsModifier func(options *metav1.ListOptions)) *cache.ListWatch {
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		optionsModifier(&options)
		return c.Endpoints(namespace).List(ctx, options)
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		optionsModifier(&options)
		return c.Endpoints(namespace).Watch(ctx, options)
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}
