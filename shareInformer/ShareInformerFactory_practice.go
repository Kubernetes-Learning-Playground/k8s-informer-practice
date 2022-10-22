package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/src"
	v11 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

type MyFactory struct {
	client *kubernetes.Clientset	// 客户端
	informers map[reflect.Type]cache.SharedIndexInformer	// 反射的informer map
}

func NewMyFactory(client *kubernetes.Clientset) *MyFactory {
	return &MyFactory{
		client: client,
		informers: make(map[reflect.Type]cache.SharedIndexInformer),
	}
}

// 支持pod informer
func (f *MyFactory) PodInformer() cache.SharedIndexInformer {

	if informer, ok := f.informers[reflect.TypeOf(&v1.Pod{})]; ok {
		return informer
	}
	podLW := cache.NewListWatchFromClient(f.client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	informer := cache.NewSharedIndexInformer(podLW, &v1.Pod{}, 0, indexers)
	f.informers[reflect.TypeOf(&v1.Pod{})] = informer
	return informer
}

// 支持deployment informer
func (f *MyFactory) DeploymentInformer() cache.SharedIndexInformer {

	if informer, ok := f.informers[reflect.TypeOf(&v11.Deployment{})]; ok {
		return informer
	}

	deploymentLW := cache.NewListWatchFromClient(f.client.AppsV1().RESTClient(), "deployments", "default", fields.Everything())
	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	}
	informer := cache.NewSharedIndexInformer(deploymentLW, &v11.Deployment{}, 0, indexers)
	f.informers[reflect.TypeOf(&v11.Deployment{})] = informer
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
	// 注册回调函数
	fact.PodInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// 注册函数
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的pod:", obj.(*v1.Pod).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("修改的pod:",oldObj.(*v1.Pod).Name, newObj.(*v1.Pod).Name)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的pod",obj.(*v1.Pod).Name)
		},
	})

	// 注册handler
	fact.DeploymentInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的deployment：",obj.(*v11.Deployment).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("更新的deployment：", oldObj.(*v11.Deployment).Name, newObj.(*v11.Deployment).Name)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的deployment：",obj.(*v11.Deployment).Name)
		},
	})


	fact.Start()


	r := gin.New()

	defer func() {
		_ = r.Run(":8082")
	}()

	fmt.Println("启动服务器")

	r.GET("/pod", func(c *gin.Context) {
		fact.PodInformer().GetIndexer().List()
		//c.JSON(200, fact.PodInformer().GetIndexer().IndexKeys(cache.NamespaceIndex, "default"))
	})

	r.GET("/deployment", func(c *gin.Context) {
		fact.DeploymentInformer().GetIndexer().List()
	})




}
