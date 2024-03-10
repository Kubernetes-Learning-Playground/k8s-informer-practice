package generics_informer

import (
	"context"
	"errors"
	"fmt"
	"k8s-informer-controller-practice/generics_informer/config"
	"k8s-informer-controller-practice/generics_informer/option"
	"k8s-informer-controller-practice/generics_informer/workqueue"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes" // import known versions
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/controller-manager/pkg/informerfactory"
	"k8s.io/klog/v2"
	"sync"
)

// GenericsCollector 收集器接口
type GenericsCollector interface {
	Run() error
	Stop()
	AddEventHandler(handler HandleFunc)
}

// genericsCollector 通用型 Informer 具体实现
type genericsCollector struct {
	// option 配置文件
	option *option.CollectorOption
	// restMapper 提供 GVR GVK 映射表
	restMapper meta.RESTMapper
	// kubeClient k8s client
	kubeClient clientset.Interface
	// workQueue 工作队列，当 informer list-watch 到资源对象后，
	// 会统一放入这个队列中给调用方调用
	workQueue workqueue.Queue
	// sharedInformers informer 对象
	sharedInformers informerfactory.InformerFactory
	// monitors 监视器，是一个 map 对象，key: GVR ,value: informer
	monitors monitors
	// monitorSets 记录有哪些 GVR 资源对象
	monitorSets map[schema.GroupVersionResource]struct{}
	// HandleFunc 调用方自定义方法
	HandleFunc HandleFunc
	// running 启动标示
	running     bool
	monitorLock sync.RWMutex
	// stopCh 停止通知 chan
	stopCh <-chan struct{}
}

func newGenericsCollector(
	option *option.CollectorOption,
	kubeClient clientset.Interface,
	monitorSet map[schema.GroupVersionResource]struct{}) (GenericsCollector, error) {

	gr, err := restmapper.GetAPIGroupResources(kubeClient.Discovery())
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	sharedInformers := informers.NewSharedInformerFactory(kubeClient, option.ResyncPeriod)

	return &genericsCollector{
		option:          option,
		kubeClient:      kubeClient,
		restMapper:      mapper,
		monitorLock:     sync.RWMutex{},
		workQueue:       workqueue.NewWorkQueue(option.MaxReQueueTime),
		running:         false,
		sharedInformers: sharedInformers,
		monitors:        map[schema.GroupVersionResource]*monitor{},
		monitorSets:     monitorSet,
	}, nil
}

func NewGenericsCollectorWithGVR(option *option.CollectorOption,
	kubeClient clientset.Interface, gvr ...string) (GenericsCollector, error) {
	monitorSet, err := newGVR(gvr...)
	if err != nil {
		return nil, err
	}
	return newGenericsCollector(option, kubeClient, monitorSet)
}

// NewGenericsCollectorWithConfig
// input: 配置文件路径, 配置项, k8s 客户端
func NewGenericsCollectorWithConfig(path string, option *option.CollectorOption,
	kubeClient clientset.Interface) (GenericsCollector, error) {
	monitorSet, err := newGVRFromConfig(path)
	if err != nil {
		return nil, err
	}
	return newGenericsCollector(option, kubeClient, monitorSet)
}

// monitor 监视器
type monitor struct {
	// controller informer 对象
	controller cache.Controller
	// store 本地缓存
	store cache.Store

	stopCh chan struct{}
}

// AddEventHandler 调用方自定义方法
func (gc *genericsCollector) AddEventHandler(handler HandleFunc) {
	gc.HandleFunc = handler
}

type HandleFunc func(object *workqueue.QueueResource) error

type monitors map[schema.GroupVersionResource]*monitor

func (gc *genericsCollector) resyncMonitors(deletableResources map[schema.GroupVersionResource]struct{}) error {
	if err := gc.initMonitorSet(deletableResources); err != nil {
		return err
	}
	gc.startMonitors()
	return nil
}

func (gc *genericsCollector) controllerFor(resource schema.GroupVersionResource, kind schema.GroupVersionKind) (cache.Controller, cache.Store, error) {
	// 处理方法, 核心逻辑就是放入工作队列中
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if o, ok := obj.(runtime.Object); ok {
				qr := &workqueue.QueueResource{Object: o, EventType: workqueue.AddEvent}
				gc.workQueue.Push(qr)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if o, ok := newObj.(runtime.Object); ok {
				qr := &workqueue.QueueResource{Object: o, EventType: workqueue.UpdateEvent}
				gc.workQueue.Push(qr)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if o, ok := obj.(runtime.Object); ok {
				qr := &workqueue.QueueResource{Object: o, EventType: workqueue.DeleteEvent}
				gc.workQueue.Push(qr)
			}
		},
	}
	shared, err := gc.sharedInformers.ForResource(resource)
	if err != nil {
		klog.Infof("unable to use a shared informer for resource %q, kind %q: %v", resource.String(), kind.String(), err)
		return nil, nil, err
	}
	klog.Infof("using a shared informer for resource %q, kind %q", resource.String(), kind.String())
	shared.Informer().AddEventHandlerWithResyncPeriod(handlers, gc.option.ResyncPeriod)
	return shared.Informer().GetController(), shared.Informer().GetStore(), nil
}

func (gc *genericsCollector) initMonitorSet(resources map[schema.GroupVersionResource]struct{}) error {
	gc.monitorLock.Lock()
	defer gc.monitorLock.Unlock()
	errList := make([]error, 0)

	for resource := range resources {
		kind, err := gc.restMapper.KindFor(resource)
		if err != nil {
			return err
		}
		c, s, err := gc.controllerFor(resource, kind)
		if err != nil {
			errList = append(errList, err)
		}
		gc.monitors[resource] = &monitor{store: s, controller: c}
	}

	if len(errList) != 0 {
		return fmt.Errorf("controller init error: %s", errList)
	}
	return nil
}

func newGVRFromConfig(path string) (map[schema.GroupVersionResource]struct{}, error) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	m := map[schema.GroupVersionResource]struct{}{}
	for _, v := range cfg.Resources.GVR {
		groupVersionResource, err := config.ParseIntoGvr(v, "/")
		if err != nil {
			return nil, err
		}
		m[groupVersionResource] = struct{}{}
	}
	return m, err
}

func newGVR(gvr ...string) (map[schema.GroupVersionResource]struct{}, error) {
	m := map[schema.GroupVersionResource]struct{}{}
	for _, v := range gvr {
		groupVersionResource, err := config.ParseIntoGvr(v, "/")
		if err != nil {
			return nil, err
		}
		m[groupVersionResource] = struct{}{}
	}
	return m, nil
}

// startMonitors 启动监视器
func (gc *genericsCollector) startMonitors() {
	gc.monitorLock.Lock()
	defer gc.monitorLock.Unlock()

	if !gc.running {
		return
	}

	// 遍历所有 map 中的 informer，并启动
	monitors := gc.monitors
	started := 0
	for _, monitor := range monitors {
		if monitor.stopCh == nil {
			monitor.stopCh = make(chan struct{})
			gc.sharedInformers.Start(gc.stopCh)
			go monitor.Run()
			started++
		}
	}
	klog.Infof("started %d new monitors, %d currently running", started, len(monitors))
}

func (m *monitor) Run() {
	m.controller.Run(m.stopCh)
}

func (gc *genericsCollector) isSynced() bool {
	gc.monitorLock.Lock()
	defer gc.monitorLock.Unlock()

	if len(gc.monitors) == 0 {
		klog.Info("GenericsCollector monitor not synced: no monitors")
		return false
	}

	for resource, monitor := range gc.monitors {
		if !monitor.controller.HasSynced() {
			klog.Infof("GenericsCollector monitor not yet synced: %+v", resource)
			return false
		}
	}
	return true
}

func (gc *genericsCollector) runProcess(ctx context.Context) {
	for gc.process(ctx) {
	}
}

func (gc *genericsCollector) process(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			klog.Info("exit work queue...")
			gc.workQueue.Close()
			return true
		default:
		}

		// 不断由队列中获取元素处理
		obj, err := gc.workQueue.Pop()
		if err != nil {
			klog.Errorf("work queue pop error: %s\n", err)
			continue
		}

		if gc.HandleFunc != nil {
			// 调用方业务逻辑如果出错，会重新入队
			if err = gc.HandleFunc(obj); err != nil {
				klog.Errorf("handle obj from work queue error: %s\n", err)
				// 重新入列
				_ = gc.workQueue.ReQueue(obj)
			} else {
				// 完成就结束
				gc.workQueue.Finish(obj)
			}
		}

	}
}

func (gc *genericsCollector) Run() error {
	return gc.RunWithContext(context.TODO())
}

func (gc *genericsCollector) RunWithContext(ctx context.Context) error {
	klog.Infof("GenericsCollector running")
	defer klog.Infof("GenericsCollector stopping")
	defer gc.workQueue.Close()

	// 1. 设置 stop channel
	gc.monitorLock.Lock()
	gc.running = true
	gc.monitorLock.Unlock()

	go gc.process(ctx)

	err := gc.initMonitorSet(gc.monitorSets)
	if err != nil {
		return err
	}
	gc.startMonitors()

	// 2. 等待 informers 的 cache 同步完成
	if !cache.WaitForCacheSync(gc.stopCh, gc.isSynced) {
		return errors.New("wait for cache sync error")
	}
	<-gc.stopCh
	return nil
}

func (gc *genericsCollector) Stop() {

	gc.monitorLock.Lock()
	defer gc.monitorLock.Unlock()
	monitors := gc.monitors
	stopped := 0
	for _, monitor := range monitors {
		if monitor.stopCh != nil {
			stopped++
			close(monitor.stopCh)
		}
	}

	gc.monitors = nil
	klog.Infof("stopped %d of %d monitors", stopped, len(monitors))
}
