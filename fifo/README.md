## DeltaFIFO队列原理简介与使用事例
DeltaFIFO队列在informer机制中如下图所示，它作为了**Apiserver**和**本地**(**Indexer**、**Listener**)之间的桥梁。它是一个生产者消费者队列，拥有FIFO的特性，操作的资源对象为Delta。

每一个Delta包含一个操作类型和操作对象。
1. FiFo:先进先出队列 有队列基本的方法(ADD UPDATE DELETE LIST POP CLOSE)
2. Delta: 存储**对象**与**对象的行为** Added Updated Deleted Sync
![]()

下面可视化DeltaQueue中最主要的两个存储结构queue和items。
![]()

###queue
存储key，对于key的生成方式keyOf，默认是取obj的namespace/name，若namespace为空，即直接为name。
是“有序”的，用来提供DeltaFIFO中FIFO的特性
与items中的key一一对应(正常情况下queue与items数量不多不少，刚好对应)
其中的key都是唯一的(在函数queueActionLocked中实现，该函数向DeltaFIFO添加元素)
###items
key与queue中key的生成方式一致
values中存储的为Deltas数组，同时保证其中必须至少有一个Delta
每一个Delta中包含:Type(操作类型)和Obj(对应的对象)，Type的类型如下
###Type的类型
Added ：增加
Updated：更新
Deleted：删除
Replaced：重新list(relist)，这个状态是由于watch event出错，导致需要进行relist来进行全盘同步。需要设置EmitDeltaTypeReplaced=true才能显示这个状态，否为默认为Sync。
Sync：本地同步

## PUSH操作

## POP操作

