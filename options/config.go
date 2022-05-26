package options

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	ManagerOptions    *ManagerOptions    `json:"manager,omitempty" yaml:"manager,omitempty" mapstructure:"manager,omitempty"`
	KubernetesOptions *KubernetesOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes,omitempty"`

	// todo-add 使用到的各控制器的options
	FalconOptions *FalconOptions `json:"falcon,omitempty" yaml:"falcon,omitempty" mapstructure:"falcon,omitempty"`
}

func TryLoadFromDisk(configFile string) (*Config, error) {
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, err
	}
	dir, file := filepath.Split(configFile)
	fileType := filepath.Ext(file)
	viper.AddConfigPath(dir)
	viper.SetConfigName(strings.TrimSuffix(file, fileType))
	viper.SetConfigType(strings.TrimPrefix(fileType, "."))

	viper.SetEnvPrefix("kubetortoise")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, errors.Errorf("解析配置文件错误:%s", err.Error())
		}
	}
	conf := NewDefaultConfig()
	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, err
}
func NewDefaultConfig() *Config {
	return &Config{
		ManagerOptions:    NewDefaultManagerOptions(),
		KubernetesOptions: NewDefaultKubernetesOptions(),
		FalconOptions:     NewDefaultFalconOptions(),
	}
}

func (c *Config) Validate() []error {
	var errs []error
	// todo add options validate functions
	errs = append(errs, c.ManagerOptions.Validate()...)
	errs = append(errs, c.KubernetesOptions.Validate()...)
	errs = append(errs, c.FalconOptions.Validate()...)
	return errs
}
