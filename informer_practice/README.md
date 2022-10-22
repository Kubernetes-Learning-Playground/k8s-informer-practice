## k8s控制器工作机制

### 控制器：
**1. 声明式定义：保持实际状态与期望状态的一致性。**

**2. informer()与list()机制同时使用**

### informer机制
主要用于接收监听资源对象(可以是调用k8s api也可以是自定义crd资源)的变化事件(Event)，

并针对不同事件(add update delete)进行不同的回调方法。

当Get或List资源时，informer不会直接去请求api-server，而是查找缓存在本地内存的数据。
1. 查询快 
2. 减少对api-server压力

**informer机制组件与步骤**
1. **Reflector**：用于监听k8s资源对象，主要使用list-watch方法进行。
   
   a. Reflector将资源版本号设为0，使用list获得对象资源（初始化资源）。
   
   b. 通过watch方法监听api server，当资源版本号有变化，就更新数据，并放入delta fifo中，使本地缓存与etcd中的保持一致。
   
   c. resync：内部有同步机制，会周期的更新list操作，不会只在初始化的时候list而已。
2. **delta fifo**：用于入队与出队操作(类似一个生产者与消费者模型)
   a. 将对象加入本地缓存store中。
   b. 触发事先定义好的handler回调函数。
3. **indexer**：根据多个索引函数来维护加入的索引。
   a. 使用MetaNamespaceKeyFunc 函数 key为 ns value 为

### list()机制
主动查询资源对象的list接口。




