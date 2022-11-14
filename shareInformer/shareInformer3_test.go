package main

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"testing"
)

func TestShareInformer3(t *testing.T) {
	//
	client := src.InitClient()

	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace("default"))
	deploymentInformer := informerFactory.Apps().V1().Deployments()
	informer := deploymentInformer.Informer()
	deploymentList := deploymentInformer.Lister()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	stopC := make(chan struct{})
	defer close(stopC)

	informerFactory.Start(stopC)
	informerFactory.WaitForCacheSync(stopC)

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


