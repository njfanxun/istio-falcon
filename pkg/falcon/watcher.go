package falcon

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchTools "k8s.io/client-go/tools/watch"
)

func (m *Manager) RunWatcher(ctx context.Context) error {
	var err error
	m.retryWatcher, err = watchTools.NewRetryWatcher("1", &cache.ListWatch{
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return m.istioClientset.NetworkingV1alpha3().Gateways(v1.NamespaceAll).Watch(context.Background(), metaV1.ListOptions{
				TypeMeta: metaV1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: "networking.istio.io/v1alpha3",
				},
			})
		},
	})

	if err != nil {
		return errors.Errorf("creating gateway watcher error:%s", err.Error())
	}
	ch := m.retryWatcher.ResultChan()
	logrus.Infoln("Beginning watching istio Gateway in all namespaces")
	for event := range ch {
		switch event.Type {
		case watch.Added, watch.Modified, watch.Deleted:
			if IsGatewayEventObject(event) {
				m.ReloadGateway(ctx)
			}
		case watch.Bookmark:
		case watch.Error:
			logrus.Error("Error attempting to watch Kubernetes Nodes")
			errObject := errors2.FromObject(event.Object)
			statusErr, ok := errObject.(*errors2.StatusError)
			if !ok {
				logrus.Errorf("Received an error which is not *metaV1.Status but %+v", event.Object)

			}
			status := statusErr.ErrStatus
			logrus.Errorf("%v", status)
		}
	}
	return nil
}

func (m *Manager) ReloadGateway(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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

func IsGatewayEventObject(event watch.Event) bool {

	gw, ok := event.Object.(*v1alpha3.Gateway)
	if !ok {
		logrus.Errorf("Unable to parse Gateway from watcher:%+v", event.Object)
		return false
	}
	var ps []string
	for _, server := range gw.Spec.Servers {
		if port := server.GetPort(); port != nil {
			ps = append(ps, cast.ToString(port.GetNumber()))
		}
	}
	logrus.Infof("%s Gateway[%s] -- %s", event.Type, gw.Name, strings.Join(ps, ","))
	return true
}
