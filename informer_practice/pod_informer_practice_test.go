package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"testing"
	"time"
)

func TestPodInformer(t *testing.T) {
	client := src.InitClient()
	stopC := make(chan struct{})
	defer close(stopC)

	sharedInformers := informers.NewSharedInformerFactory(client, time.Minute)
	informer := sharedInformers.Core().V1().Pods().Informer()

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

	informer.AddEventHandler(podHandler)
	fmt.Println("pod informer start!")
	informer.Run(stopC)

}
