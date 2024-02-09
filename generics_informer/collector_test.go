package generics_informer

import (
	"fmt"
	"k8s-informer-controller-practice/config"
	"k8s-informer-controller-practice/generics_informer/option"
	"k8s-informer-controller-practice/generics_informer/workqueue"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"testing"
	"time"
)

func TestGenericsCollector(t *testing.T) {
	gc, err := NewGenericsCollector(
		"./config.yaml",
		option.NewCollectorOption(5, 0),
		config.InitClientOrDie(),
	)
	if err != nil {
		klog.Fatal(err)
	}

	gc.AddEventHandler(func(object *workqueue.QueueResource) error {

		if o, ok := object.Object.(*appsv1.Deployment); ok {
			fmt.Println("dep event: ", object.EventType)
			fmt.Println("dep name: ", o.Name)
		}

		if o, ok := object.Object.(*v1.Pod); ok {
			fmt.Println("pod event: ", object.EventType)
			fmt.Println("pod name: ", o.Name)
		}
		if o, ok := object.Object.(*appsv1.DaemonSet); ok {
			fmt.Println("daemonset event: ", object.EventType)
			fmt.Println("daemonset name: ", o.Name)
		}
		return nil
	})

	err = gc.Run()
	if err != nil {
		klog.Fatal(err)
	}

	select {
	case <-time.After(time.Second * 10):
		gc.Stop()
	}

}
