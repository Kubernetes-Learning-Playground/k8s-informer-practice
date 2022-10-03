package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

type MyFactory struct {
	client *kubernetes.Clientset
	informers map[reflect.Type]cache.SharedIndexInformer
}

func NewMyFactory(client *kubernetes.Clientset) *MyFactory {
	return &MyFactory{
		client: client,
		informers: make(map[reflect.Type]cache.SharedIndexInformer),
	}
}

func (f *MyFactory) PodInformer() cache.SharedIndexInformer {

	if informer, ok := f.informers[reflect.TypeOf(&v1.Pod{})]; ok {
		return informer
	}
	podLW := cache.NewListWatchFromClient(f.client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	informer := cache.NewSharedIndexInformer(podLW, &v1.Pod{}, 0, indexers)
	f.informers [reflect.TypeOf(&v1.Pod{})] = informer
	return informer
}

func (f *MyFactory) Start() {
	ch := wait.NeverStop
	for _, i := range f.informers {
		go func(informer cache.SharedIndexInformer) {
			informer.Run(ch)
		}(i)
	}
}

func main() {
	client := src.InitClient()
	fact := NewMyFactory(client)
	fact.PodInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println(obj.(*v1.Pod).Name)
		},
	})
	fact.Start()


	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		fact.PodInformer().GetIndexer().List()
		//c.JSON(200, fact.PodInformer().GetIndexer().IndexKeys(cache.NamespaceIndex, "default"))
	})
	r.Run("8082")


}
