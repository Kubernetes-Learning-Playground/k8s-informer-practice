package informer_practice

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/config"
	"k8s-informer-controller-practice/informer_practice/factory"
	metav1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"testing"
)

/*
	http://xxxxxxxxxx:8082/apps/v1/deployments?labels[app]=webapp&annotations[deployment.kubernetes.io/revision]=1
	http://xxxxxxxxxx:8082/core/v1/pods
	http://xxxxxxxxxx:8082/core/v1/configmaps


*/

var (
	registerMap = map[string]struct{}{}
)

func TestConfigMapInformerWithGin(t *testing.T) {

	client := config.InitClient()

	fact := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace("default"))

	podInformer := factory.Watch(fact, "v1", "pods")
	registerMap["pods"] = struct{}{}
	err := podInformer.AddIndexers(cache.Indexers{
		"labels":      factory.LabelIndexFunc, // 除了label外，也可以加入自定义检索的index
		"annotations": factory.AnnotationsIndexFunc,
	})
	if err != nil {
		panic(err)
	}

	configMapInformer := factory.Watch(fact, "v1", "configmaps")
	registerMap["configmaps"] = struct{}{}
	err = configMapInformer.AddIndexers(cache.Indexers{
		"labels":      factory.LabelIndexFunc, // 除了label外，也可以加入自定义检索的index
		"annotations": factory.AnnotationsIndexFunc,
	})
	if err != nil {
		panic(err)
	}

	deploymentInformer := factory.Watch(fact, "apps/v1", "deployments")
	registerMap["deployments"] = struct{}{}
	err = deploymentInformer.AddIndexers(cache.Indexers{
		"labels":      factory.LabelIndexFunc, // 除了label外，也可以加入自定义检索的index
		"annotations": factory.AnnotationsIndexFunc,
	})
	if err != nil {
		panic(err)
	}

	factMap := factory.NewMyFactory()
	factMap.InformersMap[reflect.TypeOf(&v1.Pod{})] = podInformer
	factMap.InformersMap[reflect.TypeOf(&v1.ConfigMap{})] = configMapInformer
	factMap.InformersMap[reflect.TypeOf(&metav1.Deployment{})] = deploymentInformer

	// 开始加入
	factory.Start(fact)

	// 服务端启动
	r := gin.New()

	defer func() {
		r.Run(":8082")
	}()

	r.GET("/configmaps", func(c *gin.Context) {
		// 法一：不使用label过滤
		//configMapList, err := fact.Core().V1().ConfigMaps().Lister().List(labels.Everything())	// 不用标签过滤的方法

		// 法二：使用label过滤
		//set := labels.SelectorFromSet(map[string]string{
		//	"app":"prod",
		//})
		//configMapList, err := fact.Core().V1().ConfigMaps().Lister().List(set)	// 使用标签过滤

		// 法三：用户自定义标签过滤
		var set map[string]string
		if labelsMap, ok := c.GetQueryMap("labels"); ok {
			set = labelsMap
		}
		configMapList, err := fact.Core().V1().ConfigMaps().Lister().List(labels.SelectorFromSet(set))

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, gin.H{"data": configMapList})
		}
	})

	/*
		通用资源的informer 路由处理
		ex: http://1.14.120.233:8082/core/v1/pods?labels[app]=webapp
		请求事例：localhost:8080/core/v1/configmaps?labels[app]=dev
		思考：localhost:8080/apps/v1/deployments 为何显示不出来也没报错：因为没有使用对应的informer监听
	*/
	r.GET("/:group/:version/:resource", func(c *gin.Context) {
		var g, v, r = c.Param("group"), c.Param("version"), c.Param("resource")
		if g == "core" {
			g = "" // 当是资源组是core  时，需要传入空字符串
		}
		// 组成GVR
		objGroupVersionResource := schema.GroupVersionResource{
			Group:    g,
			Version:  v,
			Resource: r,
		}

		// 需要先在informer 监听下，才能返回资 源。
		if !IsRegistered(r) {
			c.JSON(400, gin.H{"error": fmt.Sprintf("resources isn't informed, please add it ")})
			return
		}

		informer, err := fact.ForResource(objGroupVersionResource)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}

		var set map[string]string
		if labelsMap, ok := c.GetQueryMap("labels"); ok {
			set = labelsMap
		}

		list, err := informer.Lister().List(labels.SelectorFromSet(set))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, gin.H{"data": list})
		}

	})

	// 路由请求事例：/common/configmaps.v1.?labels=user:jiang
	//: /common/deployments.v1.apps?labels=user:jiang
	//: /common/deployments.v1.apps?labels=app:nginx&annotations=deployment.kubernetes.io/revision:7
	r.GET("/common/:gvr", func(c *gin.Context) {
		gvr, _ := schema.ParseResourceArg(c.Param("gvr"))
		informer, _ := fact.ForResource(*gvr)

		//初始化都是nil
		var labelKeys, annotationKeys []string = nil, nil
		var list []string //最终结果
		if c.Query("labels") != "" {
			labelKeys, _ = informer.Informer().GetIndexer().
				IndexKeys("labels", c.Query("labels"))
		}
		if c.Query("annotations") != "" {
			annotationKeys, _ = informer.Informer().GetIndexer().
				IndexKeys("annotations", c.Query("annotations"))
		}
		// 返回indexKey
		if labelKeys != nil && annotationKeys != nil {
			list = factory.Intersect(labelKeys, annotationKeys) // 求交集
		} else if labelKeys != nil {
			list = labelKeys
		} else {
			list = annotationKeys
		}

		// 取出对象
		//objList,_ := informer.Informer().GetIndexer().
		//	ByIndex("labels",c.Query("labels"))
		// 取出index
		//indexList,_ := informer.Informer().GetIndexer().
		//	Index("labels",c.Query("labels"))

		c.JSON(200, gin.H{"data": list})
	})

}

// IsRegistered 是否注册到表中
func IsRegistered(resource string) bool {
	for k, _ := range registerMap {
		if resource == k {
			return true
		}
	}

	return false

}
