// Package heapvalue provides functions to create heap variables inline
// Eg: struct{x *int32}{x: heapvalue.NewInt32(1337)}
package heapvalue

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

func NewInt32(v int32) *int32 {
	return &v
}

func NewInt64(v int64) *int64 {
	return &v
}

func NewFloat64(f float64) *float64 {
	return &f
}

func NewString(s string) *string {
	return &s
}

func NewHostPathType(pathType corev1.HostPathType) *corev1.HostPathType {
	return &pathType
}

func NewJSONNumber(v int64) *apiextensions.JSON {
	json := apiextensions.JSON(v)
	return &json
}
