package main

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"strings"
	"testing"
)

func newSharedInformerListWatcher(groupVersionResource string, namespace string) cache.ListerWatcher {
	res := strings.Split(groupVersionResource, "/")
	clientSet := config.InitClientOrDie()

	var client rest.Interface
	if res[0] == "" {
		client = clientSet.CoreV1().RESTClient()
	} else {
		client = clientSet.AppsV1().RESTClient()
	}
	resource := res[1]
	selector := fields.Everything()
	lw := cache.NewListWatchFromClient(client, resource, namespace, selector)

	return lw
}

func TestShareInformer2(t *testing.T) {

	groupVersionResource := "apps/deployments"
	lw := newSharedInformerListWatcher(groupVersionResource, "default")
	obj := &v1.Deployment{}

	// NewSharedInformer的用途就是：当要监听同一个对象但是想对不同特定资源进行逻辑时使用
	sharedInformer := cache.NewSharedInformer(lw, obj, 0)
	// 添加一个处理程序
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*v1.Deployment)
			if !ok {
				return
			}
			fmt.Printf("created, printing namespace: %s\n", deployment.Namespace)
		},
	})
	// 添加另一个处理程序
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*v1.Deployment)
			if !ok {
				return
			}
			fmt.Printf("created, printing name: %s\n", deployment.Name)
		},
	})

	fmt.Println("Start syncing....")

	go sharedInformer.Run(wait.NeverStop)

	<-wait.NeverStop

}
