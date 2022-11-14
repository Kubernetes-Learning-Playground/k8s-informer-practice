package src

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"k8s.io/client-go/dynamic"
)

// 配置文件
func K8sRestConfig() *rest.Config {
	//// 需要注意这里的config文件目录。
	//config, err := clientcmd.BuildConfigFromFlags("", "config")
	//if err != nil {
	//	log.Fatal(err)
	//}

	var kubeConfig *string

	if home := HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Panic(err.Error())
	}

	//var config *rest.Config
	//// 使用 ServiceAccount 创建集群配置（InCluster模式）
	//if config, err := rest.InClusterConfig(); err != nil {
	//	// 使用 KubeConfig 文件创建集群配置
	//	if config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig); err != nil {
	//		log.Panic(err.Error())
	//	}
	//}



	return config
}

// 返回初始化k8s-client
func InitClient() *kubernetes.Clientset {
	c, err := kubernetes.NewForConfig(K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func InitDynamicClient() dynamic.Interface {
	d, err := dynamic.NewForConfig(K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}

	return os.Getenv("USERPROFILE")
}