package options

import (
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/util/homedir"
)

type KubernetesOptions struct {
	KubeConfig string  `json:"kubeconfig" yaml:"kubeconfig,omitempty"`
	Master     string  `json:"master,omitempty" yaml:"master,omitempty"`
	QPS        float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	Burst      int     `json:"burst,omitempty" yaml:"burst,omitempty"`
}

func NewDefaultKubernetesOptions() *KubernetesOptions {
	option := &KubernetesOptions{
		QPS:    1e6,
		Burst:  1e6,
		Master: "",
	}
	homePath := homedir.HomeDir()
	if homePath == "" {
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}
	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfig = userHomeConfig
		return option
	}
	adminConfig := path.Join("/etc/kubernetes", "admin.conf")
	if _, err := os.Stat(adminConfig); err == nil {
		option.KubeConfig = adminConfig
		return option
	}
	return option
}
func (k *KubernetesOptions) Validate() []error {
	var errors []error
	if k.KubeConfig != "" {
		if _, err := os.Stat(k.KubeConfig); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
