package factory

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"fmt"
)


func LabelIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	ret := []string{}
	if  meta.GetLabels() != nil {
		for k, v := range meta.GetLabels(){
			//  best:true
			ret = append(ret,fmt.Sprintf("%s:%s",k,v))
		}
	}
	fmt.Println(ret)
	return ret, nil
}


func AnnotationsIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	ret:=[]string{}
	if  meta.GetAnnotations()!=nil{
		for k,v:=range meta.GetAnnotations(){
			//  best:true
			ret=append(ret,fmt.Sprintf("%s:%s",k,v))
		}
	}

	return ret, nil
}

// Intersect 求切片的交集   select xxwhere a=xx and a=xx
func Intersect(slice1, slice2 []string) []string {
	m := make(map[string]struct{})
	nn := make([]string, 0)
	for _, v1 := range slice1 {
		m[v1] = struct{}{}
	}
	for _, v2 := range slice2 {
		if _,ok := m[v2];ok  {
			nn = append(nn, v2)
		}
	}
	return nn
}
