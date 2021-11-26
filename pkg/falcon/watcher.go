package falcon

import (
	"context"
	"strings"

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

func (m *Manager) InitWatcher() error {
	var err error
	m.rw, err = watchTools.NewRetryWatcher(m.ingressService.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return m.istioClientset.NetworkingV1alpha3().Gateways(v1.NamespaceAll).Watch(context.Background(), metaV1.ListOptions{})
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) RunWatcher(ctx context.Context) {
	logrus.Infoln("Beginning watch istio-Gateway in all namespaces")
	for event := range m.rw.ResultChan() {
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

}
func (m *Manager) ReleaseWatcher() {
	if m.rw != nil {
		m.rw.Stop()
		m.rw = nil
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
