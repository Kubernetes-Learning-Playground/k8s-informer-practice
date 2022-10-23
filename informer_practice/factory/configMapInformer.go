package factory

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"strings"
)

type MyFactory struct {
	InformersMap map[reflect.Type]cache.SharedIndexInformer
}

func NewMyFactory() *MyFactory {
	return &MyFactory{
		InformersMap: make(map[reflect.Type]cache.SharedIndexInformer),
	}
}


func Watch(fact informers.SharedInformerFactory, groupVersion, resource string) cache.SharedIndexInformer {
	gv := strings.Split(groupVersion,"/")
	var group,version string

	if len(gv)==1 {
		group = ""
		version = groupVersion
	} else {
		group = gv[0]
		if group == "core" {
			group = ""
		}
		version = gv[1]
	}
	grv := schema.GroupVersionResource{
		Group: group,
		Version: version,
		Resource: resource,
	}
	informer,err := fact.ForResource(grv)
	if err != nil {
		panic(err)
	}

	informer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{})
	return informer.Informer()


}
func Start(fact informers.SharedInformerFactory)  {
	ch := make(chan struct{})
	fact.Start(ch)
	fact.WaitForCacheSync(ch)
}
