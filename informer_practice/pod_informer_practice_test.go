package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"testing"
	"time"
)

func TestPodInformer(t *testing.T) {
	// client客户端
	client := config.InitClient()
	stopC := make(chan struct{})
	defer close(stopC)

	// sharedInformer实例
	sharedInformers := informers.NewSharedInformerFactory(client, time.Minute)
	informer := sharedInformers.Core().V1().Pods().Informer() // informer 可以add多个资源的实例

	// 回调函数
	podHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podObj := obj.(v1.Object)
			log.Printf("New pod added to store: %s", podObj.GetName())
		},

		UpdateFunc: func(oldObj, newObj interface{}) {
			newPodObj := newObj.(v1.Object)
			oldPodObj := oldObj.(v1.Object)
			log.Printf("New pod update pod %s update %s", newPodObj.GetName(), oldPodObj.GetName())
		},

		DeleteFunc: func(obj interface{}) {
			podObj := obj.(v1.Object)
			log.Printf("Pod deleted from store %s", podObj.GetName())
		},
	}
	informer.AddEventHandler(podHandler) // 加入handler！
	fmt.Println("pod informer start!")
	informer.Run(stopC)
	if !cache.WaitForCacheSync(wait.NeverStop, informer.HasSynced) {
		return
	}

}
