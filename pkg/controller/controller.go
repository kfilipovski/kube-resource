/*
Copyright (c) 2017 Kire Filipovski

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kfilipovski/kube-resource/pkg/client"
	"github.com/kfilipovski/kube-resource/pkg/spec"
)

const (
	fullName   = "resources.oracledx.com"
	maxRetries = 3
	funcKind   = "Resource"
	funcAPI    = "oracledx.com"
)

var (
	errVersionOutdated = errors.New("Requested version is outdated in apiserver")
	initRetryWaitTime  = 30 * time.Second
)

// Controller object
type Controller struct {
	clientset *rest.RESTClient
	scheme    *runtime.Scheme
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
}

// New initializes a controller object
func New(cfg *rest.Config) *Controller {

	// Create a new clientset which include our CRD schema
	cs, scheme, err := spec.NewClientSet(cfg)
	if err != nil {
		panic(err)
	}

	lw := client.ResourceClient(cs, scheme, "default").NewListWatch()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		lw,
		&spec.Resource{},
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	return &Controller{
		clientset: cs,
		scheme:    scheme,
		informer:  informer,
		queue:     queue,
	}
}

// Run starts the resource controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	fmt.Printf("Starting resource controller\n")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	// c.logger.Info("Kubeless controller synced and ready")

	// run one round of GC at startup to detect orphaned objects from the last time
	c.processAllItems()

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		// c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		// c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	fmt.Printf("change: %s/%s \n", ns, name)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		// TODO execute delete
		fmt.Printf("deleted: %s \n", obj)
		return nil
	}

	resourceObj := obj.(*spec.Resource)

	// TODO execute add/update

	fmt.Printf("updated: %v \n", resourceObj)
	return nil
}

func (c *Controller) processAllItems() error {
	if err := c.syncItems(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) syncItems() error {
	resources, err := client.ResourceClient(c.clientset, c.scheme, "default").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d resource(s)\n", len(resources.Items))

	// TODO execute sync and check for deleted
	// for _, res := range resources.Items {
	//
	// }

	return nil
}
