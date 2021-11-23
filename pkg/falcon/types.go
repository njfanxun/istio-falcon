package falcon

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	KubeConfig   = "kubeConfig"
	ServiceName  = "serviceName"
	Namespace    = "namespace"
	DefaultPorts = "ports"
)

type Config struct {
	KubeConfigPath        string
	IngressGatewayService string
	Namespace             string
	Ignore                string
	DefaultPorts          map[int32]string
}

func ParseEnvOrArgs() (*Config, error) {
	c := &Config{
		DefaultPorts: make(map[int32]string),
	}
	c.KubeConfigPath = viper.GetString(KubeConfig)
	if c.KubeConfigPath == "" {
		adminConfigPath := "/etc/kubernetes/admin.conf"
		homeConfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		switch {
		case fileExists(homeConfigPath):
			c.KubeConfigPath = homeConfigPath
		case fileExists(adminConfigPath):
			c.KubeConfigPath = adminConfigPath
		default:
			return nil, errors.Errorf("could not found kubeconfig file")
		}

	}

	c.IngressGatewayService = viper.GetString(ServiceName)
	if c.IngressGatewayService == "" {
		c.IngressGatewayService = "istio-ingressgateway"
	}

	c.Namespace = viper.GetString(Namespace)
	if c.Namespace == "" {
		c.Namespace = "istio-system"
	}

	c.Ignore = "istio-falcon.io/ignore"

	ps := viper.GetStringSlice(DefaultPorts)
	for _, p := range ps {
		c.DefaultPorts[cast.ToInt32(p)] = p
	}
	return c, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
