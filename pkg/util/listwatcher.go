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
// Option modifier is a function takes a ListOptions and modifies the consumed ListOptions. Provide customized modifier function
// to apply modification to ListOptions with a field selector, a label selector, or any other desired options.
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
