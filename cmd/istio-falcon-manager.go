package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/njfanxun/istio-falcon/apis"
	"github.com/njfanxun/istio-falcon/options"
	"github.com/njfanxun/istio-falcon/pkg/controllers/falcon"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	istioClient "istio.io/client-go/pkg/clientset/versioned"
	istioInformers "istio.io/client-go/pkg/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilErrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const defaultReSync = 600 * time.Second

func InitManagerCommand() *cobra.Command {
	var configFile string
	var command = &cobra.Command{
		Use:   "mgr",
		Short: "start k8s controller-manager for monitoring istio-ingressgateway",

		RunE: func(cmd *cobra.Command, args []string) error {
			figure.NewColorFigure("Istio Falcon", "", "green", true).
				Print()
			fmt.Println()
			cfg, err := options.TryLoadFromDisk(configFile)
			if err != nil {
				return err
			}
			if errs := cfg.Validate(); len(errs) != 0 {
				return utilErrors.NewAggregate(errs)
			}

			return Run(cfg, signals.SetupSignalHandler())
		},
		SilenceUsage: true,
	}
	fs := command.Flags()
	fs.StringVarP(&configFile, "config", "c", "/etc/falcon/config.yaml", "config file path")
	return command
}
func Run(cfg *options.Config, ctx context.Context) error {
	mgrOptions := manager.Options{
		LeaderElection:                cfg.ManagerOptions.LeaderElection,
		LeaderElectionNamespace:       cfg.ManagerOptions.LeaderElectionNamespace,
		LeaderElectionID:              cfg.ManagerOptions.LeaderElectionID,
		LeaderElectionReleaseOnCancel: true,
		LeaseDuration:                 &cfg.ManagerOptions.LeaseDuration,
		RenewDeadline:                 &cfg.ManagerOptions.RenewDeadline,
		RetryPeriod:                   &cfg.ManagerOptions.RetryPeriod,
	}
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubernetesOptions.KubeConfig)
	if err != nil {
		return err
	}
	config.QPS = cfg.KubernetesOptions.QPS
	config.Burst = cfg.KubernetesOptions.Burst
	var mgr manager.Manager
	mgr, err = manager.New(config, mgrOptions)
	if err != nil {
		return err
	}
	err = apis.AddToScheme(mgr.GetScheme())
	if err != nil {
		return err
	}

	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	istio := istioClient.NewForConfigOrDie(config)
	informerFactory := istioInformers.NewSharedInformerFactory(istio, defaultReSync)

	falconReconciler := falcon.NewFalconController(mgr.GetClient(), informerFactory, cfg.FalconOptions)
	if err := mgr.Add(falconReconciler); err != nil {
		return err
	}

	informerFactory.Start(ctx.Done())
	err = mgr.Start(ctx)
	if err != nil {
		return err
	}
	return nil

}
