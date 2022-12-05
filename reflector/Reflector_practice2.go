package main

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// 已经迁移到cachexxx目录中。

func main() {

	client := src.InitClient()
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)	// 缓存
	podListWatcher  := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"pods",
		"dafault",
		fields.Everything())

	// 默认下，只有支持一个回调函数。
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction: cache.MetaNamespaceKeyFunc,
		KnownObjects: store,	// 会存内容到缓存中，如果没有设置就delete不会有事件发生,就是没有本地缓存
	})

	/*
		reflector：
		资源的list-watcher
		资源对象
		delta fifo队列
		同步list时间: k8s list的时间
	 */
	rf := cache.NewReflector(podListWatcher, &v1.Pod{}, df, 0)
	ch := make(chan struct{})
	
	go func() {
		// 开始监听
		rf.Run(ch)
	}()
	
	for {
		// informer 不断消费队列
		_, _ = df.Pop(func(obj interface{}) error {
			for _, delta := range obj.(cache.Deltas) {

				// 遍历后需要判断回调类型，并加入store(缓存)中，不然无法取到事件。
				switch delta.Type {
				case cache.Added, cache.Sync:
					_ = store.Add(delta.Object)
				case cache.Updated:
					_ = store.Update(delta.Object)
				case cache.Deleted:
					_ = store.Delete(delta.Object)
				}

				fmt.Println(delta.Type, ":", delta.Object.(*v1.Pod).Name, ":", delta.Object.(*v1.Pod).Status.Phase)
			}
			return nil
		})
	}
	


}
