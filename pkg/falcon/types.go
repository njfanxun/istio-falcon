package falcon

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

const (
	KubeConfig  = "kubeConfig"
	ServiceName = "serviceName"
	Namespace   = "namespace"
)

type Config struct {
	KubeConfigPath        string
	IngressGatewayService string
	Namespace             string
	Ignore                string
}

func ParseEnvOrArgs() (*Config, error) {
	c := &Config{}
	c.KubeConfigPath = viper.GetString(KubeConfig)
	if c.KubeConfigPath == "" {
		c.KubeConfigPath = os.Getenv(strings.ToUpper(KubeConfig))
	}
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
		c.IngressGatewayService = os.Getenv(strings.ToUpper(ServiceName))
		if c.IngressGatewayService == "" {
			c.IngressGatewayService = "istio-ingressgateway"
		}
	}

	c.Namespace = viper.GetString(Namespace)
	if c.Namespace == "" {
		c.Namespace = os.Getenv(strings.ToUpper(Namespace))
		if c.Namespace == "" {
			c.Namespace = "istio-system"
		}
	}

	c.Ignore = "istio-falcon.io/ignore"
	return c, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
