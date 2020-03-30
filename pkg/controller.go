package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"time"
)
import appsinformers "k8s.io/client-go/informers/apps/v1"
import appslisters "k8s.io/client-go/listers/apps/v1"
import utilruntime "k8s.io/apimachinery/pkg/util/runtime"

type Controller struct {
	kubeclientset     kubernetes.Interface
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	workqueue         workqueue.RateLimitingInterface
}

func NewController(
	kubeclientset kubernetes.Interface,
	deploymentInformer appsinformers.DeploymentInformer) *Controller {
	controller := &Controller{
		kubeclientset:     kubeclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apps")}

	klog.Info("Setting up event handlers")

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(oldObj, newObj interface{}) {
			klog.Info("handler object update, obj: %s, new: %s", oldObj, newObj)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

func (c *Controller) handleObject(obj interface{}) {
	klog.Info("handler object, obj: %s", obj)
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting App controller")

	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	go wait.Until(c.runWorker, time.Second, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	deployment, err := c.deploymentsLister.Deployments(namespace).Get(name)

	klog.Infof("sync handler deployment '%s'", deployment.Spec.String())

	if err != nil {
		return err
	}

	return nil
}
