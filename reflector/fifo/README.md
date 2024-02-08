## DeltaFIFO队列原理简介与使用示例
DeltaFIFO队列在informer机制中如下图所示，它作为了**Apiserver**和**本地**(**Indexer**、**Listener**)之间的桥梁。它是一个生产者消费者队列(请分别关注**生产者**与**消费者**分别在informer机制中视哪两个组件！)，拥有FIFO的特性，操作的资源对象为Delta。

![](https://github.com/googs1025/k8s-informer-practice/blob/main/image/%E6%B5%81%E7%A8%8B%E5%9B%BE%20(2).jpg?raw=true)

每一个Delta包含一个操作类型和操作对象。
```go
// 存入delta fifo的value(包含事件类型+对象)
type Delta struct {
	Type   DeltaType
	Object interface{}
}

type Deltas []Delta
```

1. FiFo:先进先出队列 有队列基本的方法(ADD UPDATE DELETE LIST POP CLOSE)
2. Delta: 存储**对象**与**对象的行为** Added Updated Deleted Sync (注意这四种事件分别用来做什么的！)
3. keyFunc: 需要使用key的计算方法。ex: 在k8s环境中可以使用name+namespace的方式获得唯一标示。
4. knownObjects: 可以直接理解为本地缓存。ex: indexer or store组件(本质也是一个读写安全的map)
```go
type DeltaFIFO struct {
	// lock/cond protects access to 'items' and 'queue'.
	lock sync.RWMutex
	cond sync.Cond

	// `items` maps a key to a Deltas.
	// Each such Deltas has at least one Delta.
	items map[string]Deltas

	// `queue` maintains FIFO order of keys for consumption in Pop().
	// There are no duplicates in `queue`.
	// A key is in `queue` if and only if it is in `items`.
	queue []string

	// populated is true if the first batch of items inserted by Replace() has been populated
	// or Delete/Add/Update/AddIfNotPresent was called first.
	populated bool
	// initialPopulationCount is the number of items inserted by the first call of Replace()
	initialPopulationCount int

	// keyFunc is used to make the key used for queued item
	// insertion and retrieval, and should be deterministic.
	keyFunc KeyFunc

	// knownObjects list keys that are "known" --- affecting Delete(),
	// Replace(), and Resync()
	knownObjects KeyListerGetter

	// Used to indicate a queue is closed so a control loop can exit when a queue is empty.
	// Currently, not used to gate any of CRUD operations.
	closed bool

	// emitDeltaTypeReplaced is whether to emit the Replaced or Sync
	// DeltaType when Replace() is called (to preserve backwards compat).
	emitDeltaTypeReplaced bool
}
```
Delta fifo中最主要的两个存储结构queue和items。

```bigquery

```

###queue
存储对象key的队列。
存储key，对于key的生成方式keyOf，默认是取obj的namespace/name(使用MetaNamespaceKeyFunc方法)，若namespace为空，即直接为name。

其中的key都是唯一的(在函数queueActionLocked中实现，该函数向DeltaFIFO添加元素)
###items
key与queue中key的生成方式一致
values中存储的为Deltas数组，同时保证其中必须至少有一个Delta
每一个Delta中包含:Type(操作类型)和Obj(对应的对象)，Type的类型如下
###放入的Type事件类型
Added ：当监听到增加事件

Updated：当监听到更新事件

Deleted：当监听删除事件

Replaced：重新list(relist)，这个状态是由于watch event出错，导致需要进行relist来进行全盘同步。
需要设置EmitDeltaTypeReplaced=true才能显示这个状态，否为默认为Sync。

Sync：本地同步(从"本地缓存"读取数据到delta fifo中)

## PUSH操作
可以理解成由reflector list&&watch后所存放的地方。
#### 重点方法queueActionLocked
注：不管是哪种事件，都是由这个方法区分并实现的封装。可以读懂这个方法后，再去读对应的事件方法。
```bigquery
目录：tools/cache/delta_fifo.go
func (f *DeltaFIFO) queueActionLocked(actionType DeltaType, obj interface{}) error {
	// 计算key 
    id, err := f.KeyOf(obj)
    if err != nil {
        return KeyError{obj, err}
    }
    // 取到对象	
    oldDeltas := f.items[id]
    // 可以发现append进去的对象是Delta的形式   
    newDeltas := append(oldDeltas, Delta{actionType, obj})
    // 去重，对删除对象去重。
    newDeltas = dedupDeltas(newDeltas)

    if len(newDeltas) > 0 {
        // 如果不存在，append
        if _, exists := f.items[id]; !exists {
            f.queue = append(f.queue, id)
        }
    
        f.items[id] = newDeltas
        f.cond.Broadcast()
    } else {
        // This never happens, because dedupDeltas never returns an empty list
        // when given a non-empty list (as it is here).
        // If somehow it happens anyway, deal with it but complain.
        if oldDeltas == nil {
            klog.Errorf("Impossible dedupDeltas for id=%q: oldDeltas=%#+v, obj=%#+v; ignoring", id, oldDeltas, obj)
            return nil
        }
        klog.Errorf("Impossible dedupDeltas for id=%q: oldDeltas=%#+v, obj=%#+v; breaking invariant by storing empty Deltas", id, oldDeltas, obj)
        f.items[id] = newDeltas
        return fmt.Errorf("Impossible dedupDeltas for id=%q: oldDeltas=%#+v, obj=%#+v; broke DeltaFIFO invariant by storing empty Deltas", id, oldDeltas, obj)
    }
    return nil
}
```
## POP操作

```go
目录：tools/cache/delta_fifo.go
func (f *DeltaFIFO) Pop(process PopProcessFunc) (interface{}, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
    // 不断从queue中取出元素
	for {
		for len(f.queue) == 0 {
			// When the queue is empty, invocation of Pop() is blocked until new item is enqueued.
			// When Close() is called, the f.closed is set and the condition is broadcasted.
			// Which causes this loop to continue and return from the Pop().
			if f.closed {
				return nil, ErrFIFOClosed
			}

			f.cond.Wait()
		}
		// 处理queue item
		id := f.queue[0]
		f.queue = f.queue[1:]
		depth := len(f.queue)
		if f.initialPopulationCount > 0 {
			f.initialPopulationCount--
		}
		item, ok := f.items[id]
		if !ok {
			// This should never happen
			klog.Errorf("Inconceivable! %q was in f.queue but not f.items; ignoring.", id)
			continue
		}
		delete(f.items, id)
		// Only log traces if the queue depth is greater than 10 and it takes more than
		// 100 milliseconds to process one item from the queue.
		// Queue depth never goes high because processing an item is locking the queue,
		// and new items can't be added until processing finish.
		// https://github.com/kubernetes/kubernetes/issues/103789
		if depth > 10 {
			trace := utiltrace.New("DeltaFIFO Pop Process",
				utiltrace.Field{Key: "ID", Value: id},
				utiltrace.Field{Key: "Depth", Value: depth},
				utiltrace.Field{Key: "Reason", Value: "slow event handlers blocking the queue"})
			defer trace.LogIfLong(100 * time.Millisecond)
		}
		// 这里很重要，可以理解为消费者所要采取的动作。
		// 并且理解processor是如何去调用的。
		err := process(item)
		if e, ok := err.(ErrRequeue); ok {
			f.addIfNotPresent(id, item)
			err = e.Err
		}
		// Don't need to copyDeltas here, because we're transferring
		// ownership to the caller.
		return item, err
	}
}
```

