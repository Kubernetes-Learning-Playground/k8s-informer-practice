package config

import (
	"flag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

// K8sRestConfig kubeconfig配置文件
func K8sRestConfig() *rest.Config {

	var kubeConfig *string

	if home := homeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Panic(err.Error())
	}
	return config
}

// InitClientOrDie 初始化 k8s clientSet
func InitClientOrDie() kubernetes.Interface {
	c, err := kubernetes.NewForConfig(K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}

	return c
}

// InitDynamicClientOrDie 初始化 k8s dynamic client
func InitDynamicClientOrDie() dynamic.Interface {
	d, err := dynamic.NewForConfig(K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
