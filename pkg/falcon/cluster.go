package falcon

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const falconLock = "istio-falcon-lock"

func NewRestConfig(kubeConfigPath string, inCluster bool) (*rest.Config, error) {
	if inCluster {
		return rest.InClusterConfig()
	}
	c, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, errors.Errorf("error creating kubernetes client: %s", err.Error())
	}
	return c, nil
}

func (m *Manager) StartCluster() error {
	rl, err := resourcelock.New(resourcelock.LeasesResourceLock,
		m.config.Namespace,
		falconLock,
		m.k8sClientset.CoreV1(),
		m.k8sClientset.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: m.config.PodName,
		},
	)
	if err != nil {
		return errors.Errorf("lease lock error:%s", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-m.signalChan:
			cancel()
			m.GracefulShutdown()
		}
	}()
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				m.Run(ctx)
			},
			OnStoppedLeading: func() {
				m.Stop()

			},
			OnNewLeader: func(identity string) {
				logrus.Infof("Node [%s] is becomed leader of the cluster", identity)
			},
		},
		ReleaseOnCancel: true,
		Name:            "istio-falcon-manager",
	})
	return nil
}
