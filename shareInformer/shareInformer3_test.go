package main

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"
)

func TestShareInformer3(t *testing.T) {
	// 客户端
	client := config.InitClientOrDie()
	// 对deployment 监听
	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, time.Second*20, informers.WithNamespace("default"))
	deploymentInformer := informerFactory.Apps().V1().Deployments()

	// 创建informer
	informer := deploymentInformer.Informer()
	// 如果有add update delete 事件，就会回调下面的函数。
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	// 创建Lister
	deploymentList := deploymentInformer.Lister()

	stopC := make(chan struct{})
	defer close(stopC)
	// 启动 informer List&Watch。
	informerFactory.Start(stopC)
	// 等待所有缓存被同步
	informerFactory.WaitForCacheSync(stopC)

	// 从"本地缓存"获取 default 所有 deployment
	deployments, err := deploymentList.Deployments("default").List(labels.Everything())
	if err != nil {
		fmt.Print(err)
	}

	for _, deploy := range deployments {
		fmt.Printf("deployment Name:%s", deploy.Name)
	}

	<-stopC

}

func onDelete(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("new delete deployment: ", deploy.Name)
}

func onUpdate(obj interface{}, obj2 interface{}) {
	oldDeploy := obj.(*v1.Deployment)
	newDeploy := obj2.(*v1.Deployment)
	fmt.Println("update deployment: ", oldDeploy.Name, newDeploy.Name)
}

func onAdd(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("new add deployment: ", deploy.Name)
}
