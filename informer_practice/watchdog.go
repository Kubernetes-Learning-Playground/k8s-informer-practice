package informer_practice

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// 实现简单Informer机制

type WatchDog struct {
	// 外部给的
	lw *cache.ListWatch	// list-watcher
	objType runtime.Object	// k8s资源总称
	h cache.ResourceEventHandler	// 包含三种事件，add update delete

	// 构建函数内部生成
	reflector *cache.Reflector
	fifo *cache.DeltaFIFO
	store cache.Store
}

// NewWatchDog 构建函数
func NewWatchDog(lw *cache.ListWatch, objType runtime.Object, h cache.ResourceEventHandler) *WatchDog {

	store := cache.NewStore(cache.MetaNamespaceKeyFunc)

	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction: cache.MetaNamespaceKeyFunc,
		KnownObjects: store,
	})

	rf := cache.NewReflector(lw, objType, fifo, 0)

	return &WatchDog{
		lw: lw,
		objType: objType,
		h: h,
		store: store,
		fifo: fifo,
		reflector: rf,
	}


}

func (wd *WatchDog) Run() {
	ch := make(chan struct{})
	go func() {
		wd.reflector.Run(ch)
	}()
	
	for {
		wd.fifo.Pop(func(obj interface{}) error {

			for _, delta := range obj.(cache.Deltas) {
				switch delta.Type {
				case cache.Sync, cache.Added:
					_ = wd.store.Add(delta.Object)
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