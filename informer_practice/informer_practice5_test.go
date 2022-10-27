package informer_practice

import (
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"strings"
	"testing"
	"fmt"
)

func newInformerListWatcher(groupVersionResource string, namespace string) cache.ListerWatcher {
	res := strings.Split(groupVersionResource, "/")
	clientSet := src.InitClient()

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


func TestObjInformer2(t *testing.T) {

	groupVersionResource := "apps/deployments"
	lw := newInformerListWatcher(groupVersionResource, "default")
	obj := &v1.Deployment{}
	_, controller := cache.NewInformer(lw, obj, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*v1.Deployment)
			if !ok {
				return
			}
			fmt.Printf("新增deployment资源: %s\n", deployment.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment, ok := oldObj.(*v1.Deployment)
			if !ok {
				return
			}
			fmt.Printf("修改deployment资源: %s\n", deployment.Name)
		},
		DeleteFunc: func(obj interface{}) {
			deployment, ok := obj.(*v1.Deployment)
			if !ok {
				return
			}
			fmt.Printf("删除deployment资源: %s\n", deployment.Name)
		},
		
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go controller.Run(stopCh)

	<-stopCh

}
