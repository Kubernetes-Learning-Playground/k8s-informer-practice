package main

import (
	"fmt"
	"k8s-informer-controller-practice/config"

	//"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"strings"
	"time"
)

// cache.Store 用来本地存储资源状态
func newStore() cache.Store {
	metaFunc := cache.MetaNamespaceKeyFunc
	s := cache.NewStore(metaFunc)
	// 也可以使用 indexer
	//s := cache.NewIndexer(metaFunc, cache.Indexers{})
	return s
}

// cache.Queue 先进先出队列
// 注意在初始化时 store 作为 KnownObjects 参数传入其中，
// 因为在重新同步 (resync) 操作中 Reflector 需要知道当前的资源状态，
// 另外在计算变更 (Delta) 时，也需要对比当前的资源状态。
// 这个 KnownObjects 对队列，以及对 Reflector 都是只读的，用户需要自己维护好 store 的状态。
func newQueue(store cache.Store) cache.Queue {
	opt := cache.DeltaFIFOOptions{
		KnownObjects: store,
		//EmitDeltaTypeReplaced: true,
	}
	queue := cache.NewDeltaFIFOWithOptions(opt)
	return queue
}

func newListWatcher(groupVersionResource string, namespace string) cache.ListerWatcher {
	// 用来分api路径
	res := strings.Split(groupVersionResource, "/")
	clientSet := config.InitClient()

	var client rest.Interface
	if res[0] == "" {
		client = clientSet.CoreV1().RESTClient()
	} else {
		client = clientSet.AppsV1().RESTClient()
	}
	resource := res[1]
	selector := fields.Everything()
	lw := cache.NewListWatchFromClient(client, resource, namespace, selector)

	return lw

}

// newObjReflector 用于创建一个 cache.Reflector 对象，
// 当 Reflector 开始运行 (Run) 后，队列中就会推入新收到的事件。
func newObjReflector(groupVersionResource string, namespace string, obj v1.Pod, queue cache.Queue) *cache.Reflector {
	// 第 2 个参数是 expectedType, 用此参数限制进入队列的事件，
	// 当然在 List 和 Watch 操作时返回的数据就只有一种类型，这个参数只起校验的作用；
	// 第 4 个参数是 resyncPeriod，
	// 这里传了 0，表示从不重新同步（除非连接超时或者中断），
	// 如果传了非 0 值，会定期进行全量同步，避免累积和服务器的不一致，
	// 同步过程中会产生 SYNC 类型的事件。
	lw := newListWatcher(groupVersionResource, namespace)
	return cache.NewReflector(lw, &obj, queue, 0)
}

func main() {
	// 创建本地缓存
	store := newStore()
	// 创建delta fifo
	queue := newQueue(store)
	// 监听资源可以自己改变。 ex: /apps/deployments, appsv1.Deployment{}
	groupVersionResource := "/pods"
	reflector := newObjReflector(groupVersionResource, "default", v1.Pod{}, queue)

	stopCh := make(chan struct{})
	defer close(stopCh)

	// reflector 开始运行后，队列中就会推入新收到的事件
	go reflector.Run(stopCh)

	processObj := func(obj interface{}) error {
		// 最先收到的事件会被最先处理
		for _, d := range obj.(cache.Deltas) {
			switch d.Type {
			case cache.Sync, cache.Added, cache.Updated:
				// 判断是否本地缓存里有，区分add update方法
				if _, exists, err := store.Get(d.Object); err == nil && exists {
					if err := store.Update(d.Object); err != nil {
						return err
					}
				} else {
					if err := store.Add(d.Object); err != nil {
						return err
					}
				}
				// delete方法，直接在本地缓存找到后删除
			case cache.Deleted:
				if err := store.Delete(d.Object); err != nil {
					return err
				}
			}
			pod, ok := d.Object.(*v1.Pod)
			//deployment, ok := d.Object.(*metav1.Deployment)
			if !ok {
				return fmt.Errorf("not config: %T", d.Object)
			}
			fmt.Printf("事件类型：%s, 资源名称： %s\n", d.Type, pod.Name)
		}
		return nil
	}

	fmt.Println("Start syncing...")
	// 持续运行直到 stopCh 关闭
	wait.Until(func() {
		for {
			_, err := queue.Pop(processObj) // 由delta fifo中取出对象，并使用定义好的processObj 进行消费操作
			if err != nil {
				panic(err)
			}
		}
	}, time.Second, stopCh)

}
