package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"strings"
	"testing"
)

func newInformerListWatcher(groupVersionResource string, namespace string) cache.ListerWatcher {
	res := strings.Split(groupVersionResource, "/")
	clientSet := config.InitClient()

	var client rest.Interface

	resource := res[1]
	// 使用工厂方法初始化客户端
	client, err := ClientFactory(res[0], clientSet)
	if err != nil {
		log.Fatal(err)
	}
	selector := fields.Everything()
	lw := cache.NewListWatchFromClient(client, resource, namespace, selector)

	return lw
}

// ClientFactory 使用工厂方法扩展，可以自行扩展。
func ClientFactory(group string, client kubernetes.Interface) (rest.Interface, error) {

	// 可以自己扩展
	switch group {
	case corev1.GroupName:
		return client.CoreV1().RESTClient(), nil
	case appsv1.GroupName:
		return client.AppsV1().RESTClient(), nil
	case batchv1.GroupName:
		return client.BatchV1().RESTClient(), nil
	case eventv1.GroupName:
		return client.EventsV1().RESTClient(), nil
	case nodev1.GroupName:
		return client.NodeV1().RESTClient(), nil
	case rbacv1.GroupName:
		return client.RbacV1().RESTClient(), nil
	}

	return nil, fmt.Errorf("no find %s name clientset", group)
}

func TestObjInformer2(t *testing.T) {

	// 这里可以根据需要修改
	namespace := "default"
	groupVersionResource := "rbac.authorization.k8s.io/roles"

	lw := newInformerListWatcher(groupVersionResource, namespace)
	//obj := &appsv1.Deployment{}
	obj := &rbacv1.Role{}
	// 注意这里obj与反射的地方都需要修改
	_, controller := cache.NewInformer(lw, obj, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*rbacv1.Role)
			if !ok {
				return
			}
			fmt.Printf("新增deployment资源: %s\n", deployment.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment, ok := oldObj.(*rbacv1.Role)
			if !ok {
				return
			}
			fmt.Printf("修改deployment资源: %s\n", deployment.Name)
		},
		DeleteFunc: func(obj interface{}) {
			deployment, ok := obj.(*rbacv1.Role)
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
