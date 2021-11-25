package falcon

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	watchTools "k8s.io/client-go/tools/watch"
)

type Manager struct {
	k8sClientset   *kubernetes.Clientset
	istioClientset *versioned.Clientset
	signalChan     chan os.Signal
	config         *Config
	ingressService *v1.Service
	rw             *watchTools.RetryWatcher
	mutex          sync.Mutex
}

func NewManager(c *Config) (*Manager, error) {
	mgr := &Manager{
		config:     c,
		signalChan: make(chan os.Signal, 1),
		mutex:      sync.Mutex{},
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
	signal.Notify(mgr.signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)
	return mgr, nil
}

func (m *Manager) Start() {
	logrus.Info("Starting Istio-Falcon Manager...")

	m.InitIngressService()

	logrus.Info("Finished initialize istio-ingressgateway service")

	err := m.InitWatcher()
	if err != nil {
		logrus.Errorf("Creating Gateway watcher error:%s", err.Error())
		return
	}
	logrus.Info("Finished initialize Gateway Resource watcher")

	err = m.StartCluster()
	if err != nil {
		logrus.Error(err)
	}

}

// UpdateService /** @Description: 更新k8s service */
func (m *Manager) UpdateService(ctx context.Context, service *v1.Service) error {

	svc, err := m.k8sClientset.CoreV1().Services(service.Namespace).Update(ctx, service, metaV1.UpdateOptions{})
	if err != nil {
		return err
	}
	m.ingressService = svc.DeepCopy()
	return nil
}

func (m *Manager) InitIngressService() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 获取 istio-ingressgateway service
	_ = retry.Retry(func(attempt uint) error {
		var err error
		m.ingressService, err = m.k8sClientset.CoreV1().Services(m.config.ServiceNamespace).Get(ctx, m.config.IngressGatewayService, metaV1.GetOptions{})
		if err != nil {
			logrus.Errorf("attempt[%d],%s", attempt, err.Error())
			return err
		}
		return nil
	}, strategy.Wait(5*time.Second))
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
	close(m.signalChan)
	logrus.Info("Graceful shutdown Istio-Falcon manager")
}

func (m *Manager) ReloadGateway(ctx context.Context) {

	// 修改的情况复杂，重新请求计算
	gateways, err := m.istioClientset.NetworkingV1alpha3().Gateways(metaV1.NamespaceAll).List(ctx, metaV1.ListOptions{})
	if err != nil {
		logrus.Errorf("list gateways error:%s", err.Error())
		return
	}
	var gatewayPorts map[int32]*v1.ServicePort = make(map[int32]*v1.ServicePort)
	for _, gw := range gateways.Items {
		if !m.isIgnoreGateway(&gw) {
			for _, server := range gw.Spec.Servers {
				if port := server.GetPort(); port != nil {
					p := int32(port.GetNumber())
					name := strings.ToLower(gw.GetName() + "-" + port.GetName() + "-" + port.GetProtocol())
					gatewayPorts[p] = &v1.ServicePort{
						Name:     name,
						Protocol: v1.ProtocolTCP,
						Port:     p,
					}
				}
			}
		}
	}
	// 按照现在的gateway重新生成ingressService
	newIngressService := m.ingressService.DeepCopy()
	newIngressService.Spec.Ports = newIngressService.Spec.Ports[:0] // 清空ports

	for _, port := range m.ingressService.Spec.Ports {
		if m.isDefaultPort(port.Port) {
			newIngressService.Spec.Ports = append(newIngressService.Spec.Ports, port)
			continue
		}
		// 已经开放的gateway，继续保留
		_, found := gatewayPorts[port.Port]
		if found {
			delete(gatewayPorts, port.Port)
			newIngressService.Spec.Ports = append(newIngressService.Spec.Ports, port)
		}
	}
	for _, port := range gatewayPorts {
		if !m.isDefaultPort(port.Port) {
			newIngressService.Spec.Ports = append(newIngressService.Spec.Ports, *port)
		}
	}

	err = m.UpdateService(ctx, newIngressService)
	if err != nil {
		logrus.Errorf("Update istio-ingressgateway service error:%s", err.Error())
	}

}
