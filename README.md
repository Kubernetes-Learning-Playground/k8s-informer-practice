## k8s-informer机制开发

### 项目主要思路：
1. 练习运用**k8s+client-go**的informer机制搞一些简单的demo。

2. 总结一些informer组件的原理逻辑。

### 
```bigquery
├── README.md
├── cachexxx // 
├── cachexxxxTest.go    // 测试用
├── fifo    // delta-fifo队列练习
│   ├── README.md
│   └── fifo_practice.go
├── go.mod
├── go.sum
├── image
├── informer_practice //自定义informer
│   ├── README.md
│   ├── WatchExamlpe.go
│   └── WatchExamlpe_test.go
├── reflector   // reflector机制练习
│   ├── README.md
│   ├── Reflector_practice.go
│   └── Reflector_practice2.go
├── shareInformer   // shareInformer机制练习
│   ├── ShareInformerFactory_practice.go
│   └── ShareInformer_practice.go
└── src // client初始化
    └── K8sConfig.go
```


