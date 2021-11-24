package falcon

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/watch"
)

type Manager struct {
	k8sClientset   *kubernetes.Clientset
	istioClientset *versioned.Clientset
	signalChan     chan os.Signal
	config         *Config
	ingressService *v1.Service
	retryWatcher   *watch.RetryWatcher
	mutex          sync.Mutex
}

func NewManager(c *Config) (*Manager, error) {
	mgr := &Manager{
		config:     c,
		signalChan: make(chan os.Signal, 1),
	}

	config, err := NewRestConfig(c.KubeConfigPath, c.InCluster)
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
		return
	}
	m.StartCluster()

}

// UpdateService /** @Description: 更新k8s service */
func (m *Manager) UpdateService(ctx context.Context, service *v1.Service) error {
	logrus.Info("will update service")
	svc, err := m.k8sClientset.CoreV1().Services(service.Namespace).Update(ctx, service, metaV1.UpdateOptions{})
	if err != nil {
		return err
	}
	m.ingressService = svc.DeepCopy()
	return nil
}

func (m *Manager) InitIngressService() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 获取 istio-ingressgateway service
	_ = retry.Retry(func(attempt uint) error {
		var err error
		m.ingressService, err = m.k8sClientset.CoreV1().Services(m.config.Namespace).Get(ctx, m.config.IngressGatewayService, metaV1.GetOptions{})
		if err != nil {
			logrus.Errorf("attempt[%d],%s", attempt, err.Error())
			return err
		}
		return nil
	}, strategy.Wait(5*time.Second))

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
func (m *Manager) isDefaultPort(p int32) bool {
	for port, _ := range m.config.DefaultPorts {
		if p == port {
			return true
		}
	}
	return false
}

func (m *Manager) GracefulShutdown() {
	// close(m.signalChan)
	// if m.retryWatcher != nil {
	// 	m.retryWatcher.Stop()
	// }
	logrus.Info("Closed Istio-Falcon manager...")
}
