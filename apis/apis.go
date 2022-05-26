package apis

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

var AddToSchemes runtime.SchemeBuilder

func init() {
	AddToSchemes = append(AddToSchemes, v1alpha3.SchemeBuilder.AddToScheme)
}

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
