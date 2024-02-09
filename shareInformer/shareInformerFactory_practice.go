package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/config"
	v11 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

type MyFactory struct {
	client    kubernetes.Interface                       // 客户端
	informers map[reflect.Type]cache.SharedIndexInformer // 反射的informer map
}

func NewMyFactory(client kubernetes.Interface) *MyFactory {
	return &MyFactory{
		client:    client,
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

func (f *MyFactory) ConfigMapInformer() cache.SharedIndexInformer {
	if informer, ok := f.informers[reflect.TypeOf(&v1.ConfigMap{})]; ok {
		return informer
	}

	configmapLW := cache.NewListWatchFromClient(f.client.CoreV1().RESTClient(), "configmaps", "default", fields.Everything())
	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	}
	informer := cache.NewSharedIndexInformer(configmapLW, &v1.ConfigMap{}, 0, indexers)
	f.informers[reflect.TypeOf(&v1.ConfigMap{})] = informer
	return informer
}

func (f *MyFactory) EventInformer() cache.SharedIndexInformer {
	if informer, ok := f.informers[reflect.TypeOf(&v1.Event{})]; ok {
		return informer
	}

	eventLW := cache.NewListWatchFromClient(f.client.CoreV1().RESTClient(), "events", "default", fields.Everything())
	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	}
	informer := cache.NewSharedIndexInformer(eventLW, &v1.Event{}, 0, indexers)
	f.informers[reflect.TypeOf(&v1.Event{})] = informer
	return informer
}

func (f *MyFactory) ServiceInformer() cache.SharedIndexInformer {
	if informer, ok := f.informers[reflect.TypeOf(&v1.Service{})]; ok {
		return informer
	}
	serviceLW := cache.NewListWatchFromClient(f.client.CoreV1().RESTClient(), "services", "default", fields.Everything())
	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	}
	informer := cache.NewSharedIndexInformer(serviceLW, &v1.Service{}, 0, indexers)
	f.informers[reflect.TypeOf(&v1.Service{})] = informer
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
	client := config.InitClientOrDie()
	fact := NewMyFactory(client)

	// 注册回调函数
	fact.PodInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// 注册函数
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的pod:", obj.(*v1.Pod).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("修改的pod:", oldObj.(*v1.Pod).Name, newObj.(*v1.Pod).Name)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的pod", obj.(*v1.Pod).Name)
		},
	})

	// 注册handler
	fact.DeploymentInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的deployment：", obj.(*v11.Deployment).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("更新的deployment：", oldObj.(*v11.Deployment).Name, newObj.(*v11.Deployment).Name)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的deployment：", obj.(*v11.Deployment).Name)
		},
	})

	fact.ConfigMapInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的configmap：", obj.(*v1.ConfigMap).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("更新的configmap：", oldObj.(*v1.ConfigMap).Name, newObj.(*v1.ConfigMap).Name)

		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的configmap：", obj.(*v1.ConfigMap).Name)
		},
	})

	fact.EventInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的event：", obj.(*v1.Event).Name)
			fmt.Println("新增的event Type", obj.(*v1.Event).Type)
		},
	})

	fact.ServiceInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("新增的service：", obj.(*v1.Service).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("更新的service：", oldObj.(*v1.Service).Name, newObj.(*v1.Service).Name)
			oldService := oldObj.(*v1.Service)
			newService := newObj.(*v1.Service)
			if oldService.Spec.Ports[0] != newService.Spec.Ports[0] {
				fmt.Printf("service端口有变化，由%d变为%d", oldService.Spec.Ports[0], newService.Spec.Ports[0])
			}

		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("删除的service：", obj.(*v1.Service).Name)
		},
	})

	fact.Start()

	r := gin.New()

	defer func() {
		_ = r.Run(":8082")
	}()

	fmt.Println("启动服务器")
	// TODO: 增加接口回调的结果展示
	r.GET("/pod", func(c *gin.Context) {
		fact.PodInformer().GetIndexer().List()
		//c.JSON(200, fact.PodInformer().GetIndexer().IndexKeys(cache.NamespaceIndex, "default"))
	})

	r.GET("/deployment", func(c *gin.Context) {
		fact.DeploymentInformer().GetIndexer().List()
	})

	r.GET("/configmap", func(c *gin.Context) {
		fact.ConfigMapInformer().GetIndexer().List()
	})

	r.GET("/event", func(c *gin.Context) {
		fact.EventInformer().GetIndexer().List()
	})

}
