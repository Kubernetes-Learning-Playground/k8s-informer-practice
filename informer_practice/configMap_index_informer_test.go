package informer_practice

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"log"
	"testing"
)

func TestConfigMapIndexInformer(t *testing.T) {
	// client客户端
	client := config.InitClient()
	listWatcher := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "configmaps", "default", fields.Everything()) // list
	//index := cache.Indexers{} 空的 没有建立索引与indexFunc的关系。
	// 建立index
	indexer := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, // 本来内置的index就是以namespace来当做index，所以写与不写都一样
		AnnotationsIndex:     MetaAnnotationsIndexFunc,     // 自定义index增加索引
	} // 建立索引与indexFunc的关系
	// 建立indexInformer
	myIndexer, indexInformer := cache.NewIndexerInformer(listWatcher, &v1.ConfigMap{}, 0, &ConfigMapHandler1{}, indexer)

	stopC := make(chan struct{})
	// 需要用goroutine拉资源
	go indexInformer.Run(wait.NeverStop)
	defer close(stopC)

	// 如果没有同步完毕
	if !cache.WaitForCacheSync(stopC, indexInformer.HasSynced) {
		log.Fatal("sync err")
	}
	fmt.Println(myIndexer.ListKeys()) // return: [default/kube-root-ca.crt] 切片 <namespace/资源名>

	// obj, exist, err := myIndexer.GetByKey("default/kube-root-ca.crt")
	fmt.Println(myIndexer.IndexKeys(cache.NamespaceIndex, "default"))
	fmt.Println(myIndexer.IndexKeys(AnnotationsIndex, "go"))

	select {}

}

const (
	AnnotationsIndex = "app"
)

// MetaAnnotationsIndexFunc 自定义indexFunc
func MetaAnnotationsIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}

	if app, ok := meta.GetAnnotations()["app"]; ok {
		return []string{app}, nil
	}

	return []string{}, nil
}

// ConfigMapHandler1 事件的回调函数
type ConfigMapHandler1 struct{}

func (c *ConfigMapHandler1) OnAdd(obj interface{}) {
	fmt.Println("add:", obj.(*v1.ConfigMap).Name)
}

func (c *ConfigMapHandler1) OnUpdate(oldObj, newObj interface{}) {

}

func (c *ConfigMapHandler1) OnDelete(obj interface{}) {
	fmt.Println("delete:", obj.(*v1.ConfigMap).Name)
}
