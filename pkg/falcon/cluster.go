package falcon

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (m *Manager) StartCluster() {

	signal.Notify(m.signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-m.signalChan:
			m.GracefulShutdown()
			cancel()
		}
	}()
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metaV1.ObjectMeta{
			Name:      falconLock,
			Namespace: "kube-system",
		},
		Client: m.k8sClientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: "master-1",
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				logrus.Info("OnStartedLeading")
			},
			OnStoppedLeading: func() {
				logrus.Info("OnStoppedLeading")
			},
			OnNewLeader: func(identity string) {
				logrus.Infof("Node [%s] is assuming leadership of the cluster", identity)
			},
		},
		ReleaseOnCancel: true,
	})
}
