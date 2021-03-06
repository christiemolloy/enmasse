/*
 * Copyright 2018-2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

// Code generated by informer-gen. DO NOT EDIT.

package v1beta2

import (
	time "time"

	enmassev1beta2 "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta2"
	versioned "github.com/enmasseproject/enmasse/pkg/client/clientset/versioned"
	internalinterfaces "github.com/enmasseproject/enmasse/pkg/client/informers/externalversions/internalinterfaces"
	v1beta2 "github.com/enmasseproject/enmasse/pkg/client/listers/enmasse/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MessagingAddressInformer provides access to a shared informer and lister for
// MessagingAddresses.
type MessagingAddressInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta2.MessagingAddressLister
}

type messagingAddressInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMessagingAddressInformer constructs a new informer for MessagingAddress type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMessagingAddressInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMessagingAddressInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMessagingAddressInformer constructs a new informer for MessagingAddress type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMessagingAddressInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EnmasseV1beta2().MessagingAddresses(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EnmasseV1beta2().MessagingAddresses(namespace).Watch(options)
			},
		},
		&enmassev1beta2.MessagingAddress{},
		resyncPeriod,
		indexers,
	)
}

func (f *messagingAddressInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMessagingAddressInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *messagingAddressInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&enmassev1beta2.MessagingAddress{}, f.defaultInformer)
}

func (f *messagingAddressInformer) Lister() v1beta2.MessagingAddressLister {
	return v1beta2.NewMessagingAddressLister(f.Informer().GetIndexer())
}
