package main

import (
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

/*
	1. 支持多个EventHandler，可理解为多个消费者。
	2. 内置一个Indexer(有个threadSafeMap的struct实现)
    3. 多个消费者共享了Indexer
 */


func main() {

	client := src.InitClient()
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	podLW := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())

	ss := cache.NewSharedInformer(podLW, &v1.Pod{}, 0)
	// 增加多个EventHandler，都共享同一个Reflector。
	ss.AddEventHandler()
	ss.AddEventHandler()






}
