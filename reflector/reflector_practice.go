package main

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"log"
)

func main() {

	// List namespace：是default下的所有pods
	client := config.InitClientOrDie()
	// list watcher 实现：返回一个实例，可以调用 List Watch 操作
	podLW := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	// list 操作，底层调用 client-go sdk
	list, err := podLW.List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T\n", list)
	// 反射一下，不然 interface{} 不能遍历
	podList := list.(*v1.PodList)

	for _, pod := range podList.Items {
		fmt.Println("pod Name:", pod.Name)
	}

	// Watch 操作，底层调用 client-go sdk
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
