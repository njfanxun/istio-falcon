package falcon

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
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
		case watch.Added:
			gw, ok := event.Object.(*v1alpha3.Gateway)
			if !ok {
				logrus.Errorf("Unable to parse Gateway from watcher:%+v", event.Object)
			}
			m.addGateway(ctx, gw)
		case watch.Modified:

		case watch.Deleted:

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

func (m *Manager) addGateway(ctx context.Context, gw *v1alpha3.Gateway) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.isIgnoreGateway(gw) {
		logrus.Infof("ignore gateway[%s] to expose port", gw.Name)
		return
	}
	var addList []v1.ServicePort
	for _, server := range gw.Spec.Servers {
		if port := server.GetPort(); port != nil {
			p := int32(port.GetNumber())
			_, found := m.ingressServicePorts[p]
			if !found {
				name := strings.ToLower(gw.GetName() + "-" + port.GetName() + "-" + port.GetProtocol())
				m.ingressServicePorts[p] = name
				addList = append(addList, v1.ServicePort{
					Name:     name,
					Protocol: v1.ProtocolTCP,
					Port:     p,
				})

			}
		}

	}
	if len(addList) > 0 {
		err := m.UpdateService(ctx)
		if err != nil {
			logrus.Errorf("update istio-ingressgateway service error:%s", err.Error())
			return
		}
		for _, port := range addList {
			logrus.Info(" istio-ingressgateway service expose %d port", port.Port)
		}

	} else {
		logrus.Infof("Gateway[%s] has been expose port", gw.Name)
	}

}
