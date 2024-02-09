package main

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

/*
	手动把自定义资源放入fifo中，并加入回调方法，让取出资源的时候可以执行回调函数。
*/

func newK8sDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{}
}

var (
	addNum    int
	updateNum int
	deleteNum int
)

// 简单操作delta fifo队列。

func main() {

	// delta fifo queue的作用
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{KeyFunction: cache.MetaNamespaceKeyFunc})
	// list-watch 机制拿到的资源
	// 先入先出队列

	// 1.建立对象
	dep1 := newK8sDeployment()
	dep1.Name = "dep1"
	dep1.Namespace = "default"
	_ = df.Add(dep1)

	dep2 := newK8sDeployment()
	dep2.Name = "dep2"
	dep2.Namespace = "namespace2"
	_ = df.Add(dep2)

	dep1.Name = "dep11111"
	_ = df.Update(dep1)

	fmt.Println(df.List()) // 返回所有列表
	// 3.push到fifo队列 update事件

	dep3 := newK8sDeployment()
	dep3.Name = "dep3"
	dep3.Namespace = "default"
	_ = df.Add(dep3)

	// 4.push到fifo队列 delete事件
	_ = df.Delete(dep1)

	// "不断"从fifo中pop出来。当中有个回调函数，作用是分别不同事件所有做的不同回调方法，只取出一个key的元素
	for {
		_, _ = df.Pop(func(obj interface{}, isInInitialList bool) error {
			for _, delta := range obj.(cache.Deltas) {
				fmt.Println(delta.Type, ":", delta.Object.(*appsv1.Deployment).Name, "value:", delta.Object.(*appsv1.Deployment).Namespace) // 断言为pod，因为只有pod

				// 这里进行回调，区分不同事件，可以执行业务逻辑 ex: 统计次数 加入本地缓存等操作。
				switch delta.Type {
				case cache.Added:
					fmt.Println("执行新增回调")
					addNum++
				case cache.Updated:
					fmt.Println("执行更新回调")
					updateNum++
				case cache.Deleted:
					fmt.Println("执行删除回调")
					deleteNum++
				}
			}

			return nil
		})
	}

}
