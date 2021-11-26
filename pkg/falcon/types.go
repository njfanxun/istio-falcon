package falcon

import (
	"os"
	"path/filepath"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	KubeConfig       = "kube-config"
	ServiceName      = "service-name"
	ServiceNamespace = "service-namespace"
	Namespace        = "namespace"
	DefaultPorts     = "ports"
	InCluster        = "in-cluster"
	PodName          = "POD_NAME"
)

type Config struct {
	KubeConfigPath        string
	IngressGatewayService string
	ServiceNamespace      string
	Namespace             string
	Ignore                string
	DefaultPorts          map[int32]string
	InCluster             bool
	PodName               string
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
		}

	}

	c.IngressGatewayService = viper.GetString(ServiceName)
	if c.IngressGatewayService == "" {
		c.IngressGatewayService = "istio-ingressgateway"
	}

	c.ServiceNamespace = viper.GetString(ServiceNamespace)
	if c.ServiceNamespace == "" {
		c.ServiceNamespace = "istio-system"
	}

	c.Namespace = viper.GetString(Namespace)
	if c.Namespace == "" {
		c.Namespace = "kube-system"
	}

	c.Ignore = "istio-falcon.io/ignore"

	ps := viper.GetStringSlice(DefaultPorts)
	for _, p := range ps {
		c.DefaultPorts[cast.ToInt32(p)] = p
	}

	c.InCluster = viper.GetBool(InCluster)
	c.PodName = os.Getenv(PodName)
	if c.PodName == "" {
		c.PodName = string(uuid.NewUUID())

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
