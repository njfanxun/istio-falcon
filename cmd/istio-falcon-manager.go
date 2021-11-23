package cmd

import (
	"fmt"
	"github/njfanxun/istio-falcon/pkg/falcon"

	"github.com/common-nighthawk/go-figure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitManagerCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "mgr",
		Short: "start manager for monitoring istio-ingressgateway",

		Run: func(cmd *cobra.Command, args []string) {
			figure.NewColorFigure("Istio Falcon", "", "green", true).
				Print()
			fmt.Println()
			config, err := falcon.ParseEnvOrArgs()
			if err != nil {
				logrus.Errorf("parse args error:%s", err.Error())
				return
			}

			mgr, err := falcon.NewManager(config)
			if err != nil {
				logrus.Errorf("kubeconfig can not connect k8s cluster:%s", err.Error())
				return
			}
			mgr.Start()

		},
	}
	command.PersistentFlags().String(falcon.KubeConfig, "", "k8s cluster kubeconfig file path")
	_ = viper.BindPFlag(falcon.KubeConfig, command.PersistentFlags().Lookup(falcon.KubeConfig))

	command.PersistentFlags().String(falcon.ServiceName, "istio-ingressgateway", "istio-ingressgateway service name")
	_ = viper.BindPFlag(falcon.ServiceName, command.PersistentFlags().Lookup(falcon.ServiceName))

	command.PersistentFlags().String(falcon.Namespace, "istio-system", "istio-ingressgateway service namespace")
	_ = viper.BindPFlag(falcon.Namespace, command.PersistentFlags().Lookup(falcon.Namespace))

	command.PersistentFlags().StringSlice(falcon.DefaultPorts, []string{"80", "443", "15021"}, "istio-ingressgateway service opened ports by default")
	_ = viper.BindPFlag(falcon.DefaultPorts, command.PersistentFlags().Lookup(falcon.DefaultPorts))

	return command
}
