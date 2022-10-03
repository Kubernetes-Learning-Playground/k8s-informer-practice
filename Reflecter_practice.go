package main

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"log"
)

// 练习List操作与Watch操作

func main() {

	// List namespace：是default下的所有pods
	client := src.InitClient()
	podLW := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	list, err := podLW.List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T\n", list)
	podList := list.(*v1.PodList)

	for _, pod := range podList.Items {
		fmt.Println("pod的name:", pod.Name)
	}

	// Watch 操作
	watcher, err := podLW.Watch(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case v, ok := <-watcher.ResultChan():
			if ok {
				fmt.Println(v.Type, ":", v.Object.(*v1.Pod).Name)
			}
		}
	}



}
