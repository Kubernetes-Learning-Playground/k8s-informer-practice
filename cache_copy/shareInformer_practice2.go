package cache_copy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/src"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)



type PodHandler struct {
	Msg string
}

func (p PodHandler) OnAdd(obj interface{}) {
	fmt.Println("OnAdd:"+ p.Msg, obj.(metav1.Object).GetName())
}

func (p PodHandler) OnUpdate(oldObj interface{}, newObj interface{}) {
	fmt.Println("OnUpdate:"+ p.Msg)
}

func (p PodHandler) OnDelete(obj interface{}) {
	fmt.Println("OnDelete:"+ p.Msg)
}


type MySharedInformer struct {
	// 通知
	processor *sharedProcessor
	// informer 三件套。
	reflector *Reflector
	fifo *DeltaFIFO
	store Store


}

func NewMySharedInformer(lw *ListWatch, objType runtime.Object, indexer Indexer) *MySharedInformer {

	store := NewStore(MetaNamespaceKeyFunc)
	fifo := NewDeltaFIFOWithOptions(DeltaFIFOOptions{
		KeyFunction: MetaNamespaceKeyFunc,
		KnownObjects: store,
	})

	reflector := NewReflector(lw, objType, fifo, 0)

	return &MySharedInformer{
		processor: &sharedProcessor{},
		store: store,
		fifo: fifo,
		reflector: reflector,

	}
}

// input：接口对象： 实现 OnAdd OnUpdate OnDelete方法的对象
func (msi *MySharedInformer) addEventHandler(handler ResourceEventHandler) {
	lis := newProcessListener(handler, 0, 0, time.Now(), initialBufferSize)
	msi.processor.addListener(lis)
}

// start 不断从fifo中取数据
func (msi *MySharedInformer) start(ch <-chan struct{}) {

	go func() {

		for {
			_, _ = msi.fifo.Pop(func(obj interface{}) error {

				for _, delta := range obj.(Deltas) {
					switch delta.Type {
					case Sync, Added:
						_ = msi.store.Add(delta.Object)
						msi.processor.distribute(addNotification{newObj: delta.Object}, false)
						// 用shareinformer就不能直接回调。
						//msi.h.OnAdd(delta.Object) // 实现回调
					case Deleted:
						_ = msi.store.Delete(delta.Object)
						//msi.h.OnDelete(delta.Object)
						msi.processor.distribute(deleteNotification{oldObj: delta.Object}, false)
					case Updated:
						// 更新操作，需要先get之前的资源，再更新。
						if old, exists, err := msi.store.Get(delta.Object); err == nil && exists {
							_ = msi.store.Update(delta.Object)
							//msi.h.OnUpdate(old, delta.Object)
							msi.processor.distribute(updateNotification{newObj: delta.Object, oldObj: old}, false)
						}

					}
				}
				return nil

			})
		}

	}()

	// reflector 监听资源。
	go func() {
		msi.reflector.Run(ch)
	}()
	msi.processor.run(ch)
}


// MetaLabelIndexFunc 自定义一个方法 模拟MetaNamespaceIndexFunc用的！
func MetaLabelIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	if v, ok := meta.GetLabels()["app"];ok {
		return []string{v}, nil
	}
	return []string{}, nil
}

func Test() {
	//// 未加入indexer
	//client := src.InitClient()
	//podLW := NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	//
	//
	//msi := NewMySharedInformer(podLW, &v1.Pod{})
	//// SharedInformer 可以加入多个handler
	//msi.addEventHandler(&PodHandler{})
	//msi.addEventHandler(&PodHandler{Msg: "handler1"})
	//msi.addEventHandler(&PodHandler{Msg: "handler2"})
	//msi.start(wait.NeverStop)


	// 加入indexer
	//indexers := Indexers{"namespace": MetaNamespaceIndexFunc}
	//pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
	//	Name: "pod1", Namespace: "ns1", Labels: map[string]string{"app": "l1"},
	//}}
	//pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
	//	Name: "pod2", Namespace: "ns2", Labels: map[string]string{"app": "l2"},
	//}}


	//myindex := NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers)
	//myindex.Add(pod1)
	//myindex.Add(pod2)


	//objList, _ := myindex.IndexKeys("namespace", "ns1")
	//for _, obj := range objList {
	//	//fmt.Println(myindex.GetByKey(obj))
	//}




	// 自定义MetaLableIndexFunc事例
	//indexers1 := Indexers{"app": MetaLableIndexFunc}
	//myindex1 := NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers1)
	//myindex1.Add(pod1)
	//myindex1.Add(pod2)
	//
	//
	//objList1, _ := myindex1.IndexKeys("app", "l1")
	//for _, obj := range objList1 {
	//	fmt.Println(myindex1.GetByKey(obj))
	//}


	// 结合gin框架

	indexers := Indexers{"app": MetaLabelIndexFunc}
	myindex := NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers) // 传入一个补丁函数，避免deletion时发生问题

	go func() {
		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			// 会捞出 app-nginx的pod！
			ret, _ := myindex.IndexKeys("app", "nginx")
			c.JSON(200, ret)
		})
		_ = r.Run(":8088")
	}()

	client := src.InitClient()
	podLW := NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", "default", fields.Everything())
	msi := NewMySharedInformer(podLW, &v1.Pod{}, myindex)
	msi.addEventHandler(&PodHandler{})
	msi.start(wait.NeverStop)









}


