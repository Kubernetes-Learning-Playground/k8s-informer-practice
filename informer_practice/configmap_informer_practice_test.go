package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"testing"
)

/*
	informer机制：两大功能
	1. 根据集群中的某资源的事件更新本地缓存
	2. 触发注册到informer(sharedInformer)的事件回调方法
*/

func TestConfigMapInformer(t *testing.T) {

	client := config.InitClientOrDie()
	listWatcher := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "configmaps", "default", fields.Everything())

	// informer 单一资源，只能支持一个回调
	//_, informer := cache.NewInformer(listWatcher, &v1.ConfigMap{}, 0, &ConfigMapHandler{})
	//
	//informer.Run(wait.NeverStop)

	// 使用sharedInformer 可以在单一资源中  调用多个handler回调函数
	shareInformer := cache.NewSharedInformer(listWatcher, &v1.ConfigMap{}, 0)
	shareInformer.AddEventHandler(&ConfigMapHandler{})
	shareInformer.AddEventHandler(&ConfigMap2Handler{})
	shareInformer.Run(wait.NeverStop)

	select {}

}

// 事件的回调函数
type ConfigMapHandler struct{}

func (c *ConfigMapHandler) OnAdd(obj interface{}, isInInitialList bool) {
	fmt.Println("add:", obj.(*v1.ConfigMap).Name)
}

func (c *ConfigMapHandler) OnUpdate(oldObj, newObj interface{}) {

}

func (c *ConfigMapHandler) OnDelete(obj interface{}) {
	fmt.Println("delete:", obj.(*v1.ConfigMap).Name)
}

// 事件2的回调函数
type ConfigMap2Handler struct{}

func (c *ConfigMap2Handler) OnAdd(obj interface{}, isInInitialList bool) {
	fmt.Println("add2:", obj.(*v1.ConfigMap).Name)
}

func (c *ConfigMap2Handler) OnUpdate(oldObj, newObj interface{}) {

}

func (c *ConfigMap2Handler) OnDelete(obj interface{}) {
	fmt.Println("delete2:", obj.(*v1.ConfigMap).Name)
}
