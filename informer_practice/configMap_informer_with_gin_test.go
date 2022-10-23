package informer_practice

import (
	"github.com/gin-gonic/gin"
	"k8s-informer-controller-practice/informer_practice/factory"
	"k8s-informer-controller-practice/src"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"testing"
)

func TestConfigMapInformerWithGin(t *testing.T) {

	client := src.InitClient()
	fact := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace("default"))
	configMapInformer := factory.Watch(fact, "v1", "configmaps")
	err := configMapInformer.AddIndexers(cache.Indexers{
		"labels": factory.CmIndexFunc,	// 除了label外，也可以加入自定义检索的index
		"annotations": factory.CmAnnotationsFunc,
	})
	if err != nil {
		panic(err)
	}

	factory.Start(fact)

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
			c.JSON(400, gin.H{"error":err.Error()})
		} else {
			c.JSON(200, gin.H{"data":configMapList})
		}
	})


	/*
		通用资源的informer 路由处理
		请求事例：localhost:8080/core/v1/configmaps?labels[app]=dev
		思考：localhost:8080/apps/v1/deployments 为何显示不出来也没报错
	 */
	r.GET("/:group/:version/:resource", func(c *gin.Context) {
		var g, v, r = c.Param("group"), c.Param("version"), c.Param("resource")
		if g == "core" {
			g = ""	// 当是资源组是core  时，需要传入空字符串
		}
		objGroupVersionResource := schema.GroupVersionResource{
			Group: g,
			Version: v,
			Resource: r,
		}

		informer, err := fact.ForResource(objGroupVersionResource)
		if err != nil {
			c.JSON(400, gin.H{"error":err.Error()})
		}

		var set map[string]string
		if labelsMap, ok := c.GetQueryMap("labels"); ok {
			set = labelsMap
		}



		list, err := informer.Lister().List(labels.SelectorFromSet(set))
		if err != nil {
			c.JSON(400, gin.H{"error":err.Error()})
		} else {
			c.JSON(200, gin.H{"data":list})
		}


	})


	// 路由请求事例：configmaps.v1.?labels=user:jiang
	// deployments.v1.apps?labels=user:jiang
	r.GET("/common/:gvr", func(c *gin.Context) {
		gvr,_ := schema.ParseResourceArg(c.Param("gvr"))
		informer,_ := fact.ForResource(*gvr)

		//初始化都是nil
		var labelKeys,annotationKeys []string = nil,nil
		var list []string//最终结果
		if c.Query("labels")!=""{
			labelKeys, _ = informer.Informer().GetIndexer().
				IndexKeys("labels",c.Query("labels"))
		}
		if c.Query("annotations")!=""{
			annotationKeys, _ = informer.Informer().GetIndexer().
				IndexKeys("annotations",c.Query("annotations"))
		}
		if labelKeys!=nil && annotationKeys!=nil{
			list = factory.Intersect(labelKeys, annotationKeys)//求交集
		}else if labelKeys!=nil{
			list = labelKeys
		}else{
			list = annotationKeys
		}

		// 取出对象
		//objList,_ := informer.Informer().GetIndexer().
		//	ByIndex("labels",c.Query("labels"))
		// 取出index
		//indexList,_ := informer.Informer().GetIndexer().
		//	Index("labels",c.Query("labels"))

		c.JSON(200,gin.H{"data":list})
	})


}
