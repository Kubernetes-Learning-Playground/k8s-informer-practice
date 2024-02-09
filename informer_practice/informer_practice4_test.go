package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	metav1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"strings"
	"testing"
)

// 创建控制器

func newStore() cache.Store {
	metaFunc := cache.MetaNamespaceKeyFunc
	s := cache.NewStore(metaFunc)
	return s
}

func newQueue(store cache.Store) cache.Queue {
	opt := cache.DeltaFIFOOptions{
		KnownObjects: store,
		//EmitDeltaTypeReplaced: true,
	}
	queue := cache.NewDeltaFIFOWithOptions(opt)
	return queue
}

func newObjListWatcher(groupVersionResource string, namespace string) cache.ListerWatcher {
	res := strings.Split(groupVersionResource, "/")
	clientSet := config.InitClientOrDie()

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

func newController() cache.Controller {
	groupVersionResource := "apps/deployments"
	lw := newObjListWatcher(groupVersionResource, "default")
	store := newStore()
	queue := newQueue(store)

	processObj := func(obj interface{}, isInInitialList bool) error {
		// 最先收到的事件会被最先处理
		for _, d := range obj.(cache.Deltas) {
			switch d.Type {
			case cache.Sync, cache.Added, cache.Updated:
				if _, exists, err := store.Get(d.Object); err == nil && exists {
					if err := store.Update(d.Object); err != nil {
						return err
					}
				} else {
					if err := store.Add(d.Object); err != nil {
						return err
					}
				}
			case cache.Deleted:
				if err := store.Delete(d.Object); err != nil {
					return err
				}
			}
			obj, ok := d.Object.(*metav1.Deployment)
			if !ok {
				return fmt.Errorf("not config: %T", d.Object)
			}
			fmt.Printf("事件类型：%s, 资源名称： %s\n", d.Type, obj.Name)
		}
		return nil
	}

	cfg := cache.Config{
		Queue:            queue,
		ListerWatcher:    lw,
		ObjectType:       &metav1.Deployment{},
		FullResyncPeriod: 0,
		RetryOnError:     false,
		Process:          processObj,
	}
	return cache.New(&cfg)
}

func TestObjInformer(t *testing.T) {

	controller := newController()

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go controller.Run(stopCh)

	<-stopCh

}
