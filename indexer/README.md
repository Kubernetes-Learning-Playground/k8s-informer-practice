## Indexer原理简介与使用示例

Indexer是informer机制中的本地缓存(带索引index)，可以通过读本地缓存的资源对象，减少对api-server的请求压力。
![](https://github.com/googs1025/k8s-informer-practice/blob/main/image/%E6%B5%81%E7%A8%8B%E5%9B%BE%20(6).jpg?raw=true)
#### 本地缓存Indexer接口对象
```bigquery
type Indexer interface {
    // 可以直接理解成本地缓存。
	Store
	// 在本地缓存上，增加一些index的方法  
	Index(indexName string, obj interface{}) ([]interface{}, error)
	IndexKeys(indexName, indexedValue string) ([]string, error)
	ListIndexFuncValues(indexName string) []string
	ByIndex(indexName, indexedValue string) ([]interface{}, error)
	GetIndexers() Indexers
	AddIndexers(newIndexers Indexers) error
}
```
#### 索引器的功能，需要注意一下。
```bigquery
type Index map[string]sets.String

type IndexFunc func(obj interface{}) ([]string, error)

type Indexers map[string]IndexFunc

type Indices map[string]Index
```

```bigquery
type cache struct {
    // 一个接口方法
	cacheStorage ThreadSafeStore
	// 计算func
	keyFunc KeyFunc
}
```
#### 本地缓存底层对象
```bigquery
// threadSafeMap implements ThreadSafeStore
type threadSafeMap struct {
	lock  sync.RWMutex
    // 本地缓存 是一个map
	items map[string]interface{}
    
    // 跟索引有关的属性
	// indexers maps a name to an IndexFunc
	indexers Indexers
	// indices maps a name to an Index
	indices Indices
}
```
