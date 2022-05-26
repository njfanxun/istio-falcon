package options

import (
	"github.com/pkg/errors"
)

type FalconOptions struct {
	IstioNamespace     string  `json:"namespace,omitempty" yaml:"namespace,omitempty" mapstructure:"namespace"`
	IngressgatewayName string  `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name"`
	DefaultExposePorts []int32 `json:"ports,omitempty" yaml:"ports,omitempty" mapstructure:"ports"`
	IgnoreLabel        string  `json:"ignoreLabel,omitempty" yaml:"ignoreLabel,omitempty" mapstructure:"ignoreLabel"`
}

func NewDefaultFalconOptions() *FalconOptions {
	return &FalconOptions{
		IstioNamespace:     "istio-system",
		IngressgatewayName: "istio-ingressgateway",
		DefaultExposePorts: []int32{80, 443, 15021},
		IgnoreLabel:        "networking.istio.io/gateway-open",
	}
}
func (f *FalconOptions) Validate() []error {
	var errs []error
	if f.IstioNamespace == "" {
		errs = append(errs, errors.Errorf("istio namespace must be specified"))
	}
	if f.IngressgatewayName == "" {
		errs = append(errs, errors.Errorf("istio-ingressgateway service name must be specified"))
	}
	return errs
}
