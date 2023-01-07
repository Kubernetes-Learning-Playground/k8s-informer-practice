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

/*
	sharedInformer or sharedIndexInformer
	1. 支持多个eventHandler
	2. 内置一个indexer 缓存
 */

func TestShareInformer(t *testing.T) {

	client := src.InitClient()

	// 两个方法区别
	//fact := informers.NewSharedInformerFactory(client, time.Minute)	// 这个会监听所有namespace
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

	// 可以加入多个EventHandler
	configInformer.Informer().AddEventHandler(&ConfigMapHandler{})


	// 也可以用Informer().Run(wait.NeverStop)
	//configInformer.Informer().Run(wait.NeverStop)
	// 看是启动所有的informer还是单独的。
	fact.Start(wait.NeverStop)
	// 等待所有缓存被同步
	fact.WaitForCacheSync(wait.NeverStop)
	// 如果只有一个，也可以用这个
	//if !cache.WaitForCacheSync(wait.NeverStop, configInformer.Informer().HasSynced) {
	//	return
	//}

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
