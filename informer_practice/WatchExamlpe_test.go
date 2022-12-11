package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"testing"
)

// 资源handler对象
type PodHandler struct {
}

// OnAdd 当有add事件时，使用的回调函数
func (p PodHandler) OnAdd(obj interface{}) {
	fmt.Println("OnAdd:", obj.(*v1.Pod).Name)
}

// OnUpdate 当有update事件时，使用的回调函数
func (p PodHandler) OnUpdate(oldObj, newObj interface{}) {
	fmt.Println("OnUpdate:", newObj.(*v1.Pod).Name)
}

// OnDelete 当有delete事件时，使用的回调函数
func (p PodHandler) OnDelete(obj interface{}) {
	fmt.Println("OnDelete:", obj.(*v1.Pod).Name)
}

// 加入资源handler
var _ cache.ResourceEventHandler = &PodHandler{} // 查看是否实现此接口

func TestWatchExamlpe(t *testing.T) {
	// 建立client
	client := src.InitClient()
	// 客户端
	podLW := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"pods",
		"default",
		fields.Everything(),
	)
	// 增新对象
	wd := NewWatchWatchExample(podLW, &v1.Pod{}, &PodHandler{})
	// 启动
	wd.Run()

}
