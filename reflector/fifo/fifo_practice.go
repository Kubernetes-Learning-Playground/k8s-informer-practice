package main

import (
	"fmt"
	"k8s.io/client-go/tools/cache"
)

/*
	手动把自定义资源放入fifo中，并加入回调方法，让取出资源的时候可以执行回调函数。
*/

// 自定义pod对象 (想加入的自定义资源对象。)
type pod struct {
	Name  string
	Value float64
}

// 构建pod
func newPod(name string, v float64) pod {
	return pod{
		Name:  name,
		Value: v,
	}
}

func podKeyFunc(obj interface{}) (string, error) {
	return obj.(pod).Name, nil
}

// 简单操作delta fifo队列。
// 放入对象可以是自己封装的方法，但是要确定好keyFunc。

func main() {

	// delta fifo queue的作用
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{KeyFunction: podKeyFunc})
	// list-watch 机制拿到的资源
	// 先入先出队列

	// 1.建立对象
	pod1 := newPod("pod1", 1)
	pod2 := newPod("pod2", 2)
	pod3 := newPod("pod3", 3)
	// 2.push到fifo队列 add事件
	_ = df.Add(pod1)
	_ = df.Add(pod2)
	_ = df.Add(pod3)
	fmt.Println(df.List()) // 返回所有列表
	pod1.Value = 1.111
	// 3.push到fifo队列 update事件
	_ = df.Update(pod1)
	// 4.push到fifo队列 delete事件
	_ = df.Delete(pod1)

	// 从fifo中pop出来。当中有个回调函数，作用是分别不同事件所有做的不同回调方法，只取出一个key的元素
	_, _ = df.Pop(func(obj interface{}) error {
		for _, delta := range obj.(cache.Deltas) {
			fmt.Println(delta.Type, ":", delta.Object.(pod).Name, "value:", delta.Object.(pod).Value) // 断言为pod，因为只有pod

			// 这里进行回调，区分不同事件。
			switch delta.Type {
			case cache.Added:
				fmt.Println("执行新增回调")
			case cache.Updated:
				fmt.Println("执行更新回调")
			case cache.Deleted:
				fmt.Println("执行删除回调")
			}
		}

		return nil
	})

}
