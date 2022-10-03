package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"testing"
)

type PodHandler struct {
	
}

func (p PodHandler) OnAdd(obj interface{}) {
	fmt.Println("OnAdd:", obj.(*v1.Pod).Name)
}

func (p PodHandler) OnUpdate(oldObj, newObj interface{}) {
	fmt.Println("OnUpdate:", newObj.(*v1.Pod).Name)
}

func (p PodHandler) OnDelete(obj interface{}) {
	fmt.Println("OnDelete:", obj.(*v1.Pod).Name)
}

var _ cache.ResourceEventHandler = &PodHandler{}

func WatchDogTest(t *testing.T) {
	client := src.InitClient()

	podLW := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	wd := NewWatchDog(podLW, &v1.Pod{}, &PodHandler{})
	wd.Run()

}
