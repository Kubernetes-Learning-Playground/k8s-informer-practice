## k8s-informer机制开发
![](https://github.com/googs1025/k8s-informer-practice/blob/main/image/framework.png?raw=true)
### 项目主要思路：
本项目目前不打算更新过多的源码内容，原因是网上随便一搜可以抓出一大把的源码分析博客，各路大神都已经把内容都分析的很透彻了。
但是反观练手或是demo的项目却非常少，可以说是几乎没有，有也是简单的example的代码示例。基于上述原因，设计此项目是提供更多的demo，让其他开发者可以有更多好的想法扩展informer机制。
1. 练习运用**k8s+client-go**的informer机制搞一些简单的demo。

2. 总结一些informer组件的原理逻辑。

3. 同时介绍一下informer机制里常用的组件：Reflector、Delta fifo、Indexer等



### 

- indexer [调用练习](./indexer)
- informer [调用练习](./informer_practice)
- reflector [调用练习](./reflector)
- shareInformer [调用练习](./shareInformer)
- 通用 Informer [收集器](./generics_informer)


