package main

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"log"
	"testing"
	"time"
)

func TestShareInformer(t *testing.T) {

	client := src.InitClient()

	// 两个方法区别
	//fact := informers.NewSharedInformerFactory(client, time.Minute)
	fact := informers.NewSharedInformerFactoryWithOptions(client, time.Minute, informers.WithNamespace("default"))


	// 法一：
	//fact.Core().V1().ConfigMaps().Informer()

	// 法二：
	configGVR := schema.GroupVersionResource{
		 Group: "",
		 Version: "v1",
		 Resource: "configmaps",
	}
	configInformer, err := fact.ForResource(configGVR)
	if err != nil {
		log.Println(err)
	}


	configInformer.Informer().AddEventHandler(&ConfigMapHandler{})


	fact.Start(wait.NeverStop)

	select {} // 如果不是用gin 就需要永远阻塞

}

// 事件的回调函数
type ConfigMapHandler struct {

}

func (c *ConfigMapHandler) OnAdd(obj interface{}) {
	fmt.Println("add:", obj.(*v1.ConfigMap).Name)
}

func (c *ConfigMapHandler) OnUpdate(oldObj, newObj interface{}) {

}

func (c *ConfigMapHandler) OnDelete(obj interface{}) {
	fmt.Println("delete:", obj.(*v1.ConfigMap).Name)
}
