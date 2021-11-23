package falcon

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/watch"
)

type Manager struct {
	k8sClientset        *kubernetes.Clientset
	istioClientset      *versioned.Clientset
	signalChan          chan os.Signal
	config              *Config
	ingressService      *v1.Service
	ingressServicePorts map[int32]string
	retryWatcher        *watch.RetryWatcher
	mutex               sync.Mutex
}

func NewManager(c *Config) (*Manager, error) {
	mgr := &Manager{
		config:              c,
		signalChan:          make(chan os.Signal, 1),
		ingressServicePorts: make(map[int32]string),
	}

	config, err := clientcmd.BuildConfigFromFlags("", c.KubeConfigPath)
	if err != nil {
		return nil, err
	}

	mgr.k8sClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	mgr.istioClientset, err = versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) Start() {
	logrus.Info("Starting Istio-Falcon Manager...")

	err := m.InitIngressService()
	if err != nil {
		logrus.Error(err)
	}
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return m.RunWatcher(ctx)
	})

	signal.Notify(m.signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)
	select {
	case <-m.signalChan:
		m.GracefulShutdown()
	}
}

// UpdateService /** @Description: 更新k8s service */
func (m *Manager) UpdateService(ctx context.Context) error {

	_, err := m.k8sClientset.CoreV1().Services(m.ingressService.Namespace).Update(ctx, m.ingressService, metaV1.UpdateOptions{})
	return err
}

func (m *Manager) InitIngressService() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 获取 istio-ingressgateway service

	_ = retry.Retry(func(attempt uint) error {
		var err error
		m.ingressService, err = m.k8sClientset.CoreV1().Services(m.config.Namespace).Get(ctx, m.config.IngressGatewayService, metaV1.GetOptions{})
		if err != nil {
			logrus.Infof("attempt[%d],%s", attempt, err.Error())
			return err
		}
		return nil
	}, strategy.Wait(5*time.Second))

	for _, port := range m.ingressService.Spec.Ports {
		m.ingressServicePorts[port.Port] = port.Name
	}

	logrus.Info("Finished initialize istio-ingressgateway")
	return nil
}

func (m *Manager) isIgnoreGateway(gw *v1alpha3.Gateway) bool {
	annotation, f1 := gw.GetObjectMeta().GetAnnotations()[m.config.Ignore]
	if f1 && annotation == "true" {
		return true
	}
	label, f2 := gw.GetObjectMeta().GetLabels()[m.config.Ignore]
	if f2 && label == "true" {
		return true
	}
	return false
}

func (m *Manager) GracefulShutdown() {
	close(m.signalChan)
	if m.retryWatcher != nil {
		m.retryWatcher.Stop()
	}
	logrus.Info("close istio-falcon manager...")
}
