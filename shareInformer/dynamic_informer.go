package main

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("缺少输入参数")
		return
	}

	resources := os.Args[1] // // 资源，比如 "configmaps.v1.", "deployments.v1.apps", "rabbits.v1.stable.wbsnail.com"

	dynamicClient := config.InitDynamicClientOrDie()
	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicClient, 0, "default", nil)

	// 通过 schema 包提供的 ParseResourceArg 由资源描述字符串解析出 GroupVersionResource
	resourcesGVR, _ := schema.ParseResourceArg(resources)
	if resourcesGVR == nil {
		log.Fatal("GVR解析为空")
		return
	}

	// 使用 gvr 动态生成 Informer
	informer := informerFactory.ForResource(*resourcesGVR).Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// *unstructured.Unstructured 类是所有 Kubernetes 资源类型公共方法的抽象，
		// 提供所有对公共属性的访问方法，像 GetName, GetNamespace, GetLabels 等等，
		AddFunc: func(obj interface{}) {
			s, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			fmt.Printf("created: name: %s, apiversion: %s\n", s.GetName(), s.GetAPIVersion())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldS, ok1 := oldObj.(*unstructured.Unstructured)
			newS, ok2 := newObj.(*unstructured.Unstructured)
			if !ok1 || !ok2 {
				return
			}
			fmt.Println(oldS.GetName(), newS.GetName())
			//// 要访问公共属性外的字段，可以借助 unstructured 包提供的一些助手方法：
			//oldThing, ok1, err1 := unstructured.NestedString(oldS.Object, "spec", "xxx")
			//newThing, ok2, err2 := unstructured.NestedString(newS.Object, "spec", "xxxx")
			//if !ok1 || !ok2 || err1 != nil || err2 != nil {
			//	fmt.Printf("updated: %s\n", newS.GetName())
			//}
		},
		DeleteFunc: func(obj interface{}) {
			s, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			fmt.Printf("deleted name: %s, apiversion: %s\n", s.GetName(), s.GetAPIVersion())
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go informerFactory.Start(stopCh)

	<-stopCh

}
