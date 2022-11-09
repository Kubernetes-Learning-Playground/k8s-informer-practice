package informer_practice

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// 实现简单Informer机制

// WatchExamlpe 对象
type WatchExamlpe struct {
	// 外部给的
	lw *cache.ListWatch	// list-watcher
	objType runtime.Object	// k8s资源总称，监听对象类型
	h cache.ResourceEventHandler	// 包含三种事件，add update delete

	// 构建函数内部生成
	reflector *cache.Reflector
	fifo *cache.DeltaFIFO
	store cache.Store
}

// NewWatchDog 构建函数，输入参数：lw:list-watch objType:资源种类 h:资源handler
func NewWatchWatchExample(lw *cache.ListWatch, objType runtime.Object, h cache.ResourceEventHandler) *WatchExamlpe {

	// 新建Store
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)

	// 新建FIFO
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction: cache.MetaNamespaceKeyFunc,
		KnownObjects: store,
	})

	// 新建Reflector
	rf := cache.NewReflector(lw, objType, fifo, 0)

	return &WatchExamlpe{
		lw: lw,	// list-watch
		objType: objType,	// 资源
		h: h,	// 资源的handler
		store: store,	// 本地缓存
		fifo: fifo,		// 队列
		reflector: rf,	// reflector
	}

}

func (wd *WatchExamlpe) Run() {


	ch := make(chan struct{})
	 go func() {
		// 启动run
		wd.reflector.Run(ch)
	}()

	// 不断从队列取出来
	for {
		// 从fifo队列中 pop出来
		_, _ = wd.fifo.Pop(func(obj interface{}) error {

			for _, delta := range obj.(cache.Deltas) {
				switch delta.Type {
				case cache.Sync, cache.Added:
					_ = wd.store.Add(delta.Object)	// 存入store缓存
					wd.h.OnAdd(delta.Object) // 实现回调
				case cache.Deleted:
					_ = wd.store.Delete(delta.Object)
					wd.h.OnDelete(delta.Object)
				case cache.Updated:
					// 更新操作，需要先get之前的资源，再更新。
					if old, exists, err := wd.store.Get(delta.Object); err == nil && exists {
						_ = wd.store.Update(delta.Object)
						wd.h.OnUpdate(old, delta.Object)
					}

				}
			}
			return nil

		})
	}
}