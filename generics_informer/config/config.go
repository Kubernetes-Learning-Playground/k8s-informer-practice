package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strings"
)

type Config struct {
	MaxReQueueTime int       `json:"maxRequeueTime" yaml:"maxRequeueTime"`
	Resources      Resources `json:"resources" yaml:"resources"`
}

type Resources struct {
	GVR []string `json:"gvr" yaml:"gvr"`
}

func NewConfig() *Config {
	return &Config{}
}

func loadConfigFile(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	return b
}

func LoadConfig(path string) (*Config, error) {
	config := NewConfig()
	if b := loadConfigFile(path); b != nil {

		err := yaml.Unmarshal(b, config)
		if err != nil {
			return nil, err
		}
		return config, err
	} else {
		return nil, fmt.Errorf("load config file error")
	}
}

// ParseIntoGvr 解析并指定资源对象GVR，http server 接口使用 "/" 作为分割符
// ex: "apps/v1/deployments" "core/v1/pods" "batch/v1/jobs"
func ParseIntoGvr(gvr, splitString string) (schema.GroupVersionResource, error) {
	var group, version, resource string
	gvList := strings.Split(gvr, splitString)

	// 防止越界
	if len(gvList) < 2 {
		return schema.GroupVersionResource{}, fmt.Errorf("gvr input error, please input like format apps/v1/deployments or core/v1/multiclusterresource")
	}

	if len(gvList) == 2 {
		group = ""
		version = gvList[0]
		resource = gvList[1]
	} else {
		if gvList[0] == "core" {
			gvList[0] = ""
		}
		group, version, resource = gvList[0], gvList[1], gvList[2]
	}

	return schema.GroupVersionResource{
		Group: group, Version: version, Resource: resource,
	}, nil
}
