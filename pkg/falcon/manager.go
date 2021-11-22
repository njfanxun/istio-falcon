package falcon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Manager struct {
	k8sClientset   *kubernetes.Clientset
	istioClientset *versioned.Clientset
	signalChan     chan os.Signal
	config         *Config
	ports          []*v1alpha3.Port
	ingressService *v1.Service
}

func NewManager(c *Config) (*Manager, error) {
	mgr := &Manager{
		config:     c,
		signalChan: make(chan os.Signal, 1),
		ports:      []*v1alpha3.Port{},
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
	m.initIngressGatewayService()

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

func (m *Manager) initIngressGatewayService() {
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return retry.Retry(func(attempt uint) error {
			var err error
			m.ingressService, err = m.k8sClientset.CoreV1().Services(m.config.Namespace).Get(ctx, m.config.IngressGatewayService, metav1.GetOptions{})
			if err != nil {
				logrus.Infof("not found istio-ingressgatway service,wait for attempt[%d]", attempt)
				return err
			}
			return nil
		}, strategy.Wait(5*time.Second))

	})
	err := g.Wait()
	if err != nil {
		logrus.Error(err)
	}

	if m.ingressService.Spec.Type == v1.ServiceTypeClusterIP {
		logrus.Errorf("istio-ingressgateway service is %s,not required map ports", m.ingressService.Spec.Type)

	}
}

func (m *Manager) GracefulShutdown() {
	close(m.signalChan)
	logrus.Info("close istio-falcon manager...")
}
