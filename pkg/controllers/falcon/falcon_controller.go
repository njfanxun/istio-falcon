package falcon

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/njfanxun/istio-falcon/options"
	"github.com/njfanxun/istio-falcon/pkg/utils/logger"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	v1alpha33 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioInformers "istio.io/client-go/pkg/informers/externalversions"
	"istio.io/client-go/pkg/listers/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	rt "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerName = "falcon-controller"
)

type FalconController struct {
	client.Client
	Lister         v1alpha3.GatewayLister
	Informer       cache.SharedIndexInformer
	InformerSynced cache.InformerSynced
	Queue          workqueue.Interface
	Options        *options.FalconOptions
	Logger         logr.Logger
	MaxConcurrent  int
}

func NewFalconController(client client.Client, informerFactory istioInformers.SharedInformerFactory, ops *options.FalconOptions) *FalconController {

	f := &FalconController{
		Client:         client,
		Lister:         informerFactory.Networking().V1alpha3().Gateways().Lister(),
		Informer:       informerFactory.Networking().V1alpha3().Gateways().Informer(),
		InformerSynced: informerFactory.Networking().V1alpha3().Gateways().Informer().HasSynced,
		Queue:          workqueue.NewNamed("falcon"),
		Options:        ops,
		Logger:         logger.Logger(controllerName),
		MaxConcurrent:  1,
	}

	f.Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: f.callbackFunc,
		UpdateFunc: func(oldObj, newObj interface{}) {
			f.callbackFunc(newObj)
		},
		DeleteFunc: f.callbackFunc,
	})

	return f
}

func (f *FalconController) Start(ctx context.Context) error {
	return f.Run(f.MaxConcurrent, ctx.Done())
}

func (f *FalconController) Run(workers int, stopCh <-chan struct{}) error {
	defer rt.HandleCrash()
	defer f.Queue.ShutDown()
	// 等待缓存同步完成
	if !cache.WaitForCacheSync(stopCh, f.InformerSynced) {
		return errors.Errorf("%s failed to wait for caches to sync gateways", controllerName)
	}

	for i := 0; i < workers; i++ {
		go wait.Until(f.worker, time.Second, stopCh)
	}
	<-stopCh
	return nil
}

func (f *FalconController) worker() {
	for f.processNextWorkItem() {
	}
}

func (f *FalconController) processNextWorkItem() bool {
	eKey, quit := f.Queue.Get()
	if quit {
		return false
	}
	defer f.Queue.Done(eKey)
	err := f.processIngressGatewayService()
	if err != nil {
		f.Logger.Error(err, "process ingress gateway")
	}
	return true

}

func (f *FalconController) processIngressGatewayService() error {

	// 获取所有的gateway
	gws, err := f.Lister.List(labels.Everything())
	if err != nil {
		return err
	}
	var gatewayPortMaps map[int32]string = make(map[int32]string)
	for _, gw := range gws {
		if f.IsIgnoreGateway(gw) {
			continue
		}
		for _, server := range gw.Spec.GetServers() {
			if p := server.GetPort(); p != nil {
				if f.IsEffectivePort(p.GetNumber()) {
					port := cast.ToInt32(p.GetNumber())
					protocol := strings.ToLower(p.GetProtocol())
					name, ok := gatewayPortMaps[port]
					if !ok {
						gatewayPortMaps[port] = fmt.Sprintf("%s-%d", protocol, port)
					} else {
						if !strings.Contains(name, fmt.Sprintf("%s-", protocol)) {
							gatewayPortMaps[port] = fmt.Sprintf("%s-%s", protocol, name)
						}
					}
					if len(gatewayPortMaps[port]) > 15 {
						gatewayPortMaps[port] = gatewayPortMaps[port][0:14]
					}

				}
			}
		}
	}

	var currentService corev1.Service
	err = f.Client.Get(context.Background(), types.NamespacedName{
		Namespace: f.Options.IstioNamespace,
		Name:      f.Options.IngressgatewayName,
	}, &currentService)

	if err != nil {
		f.Logger.Error(err, "not found ingress gateway service")
		return client.IgnoreNotFound(err)
	}

	updateService := currentService.DeepCopy()

	// 清空待更新的ports数组
	updateService.Spec.Ports = updateService.Spec.Ports[:0]
	var reports sets.String = sets.NewString()
	for _, sp := range currentService.Spec.Ports {
		// 属于默认开放的端口
		if lo.Contains[int32](f.Options.DefaultExposePorts, sp.Port) {
			updateService.Spec.Ports = append(updateService.Spec.Ports, sp)
			continue
		}
		name, found := gatewayPortMaps[sp.Port]
		if found {
			// gateway端口已经存在service中
			if sp.Name != name {
				sp.Name = name
				reports.Insert(fmt.Sprintf("update->%d", sp.Port))

			}
			updateService.Spec.Ports = append(updateService.Spec.Ports, sp)
			delete(gatewayPortMaps, sp.Port)
		} else {
			// gateway端口不存在service，需要删除该端口，不添加端口,标记需要更新
			reports.Insert(fmt.Sprintf("delete->%d", sp.Port))

		}
	}
	for port, name := range gatewayPortMaps {
		sp := corev1.ServicePort{
			Name:     name,
			Protocol: corev1.ProtocolTCP,
			Port:     port,
		}
		updateService.Spec.Ports = append(updateService.Spec.Ports, sp)
		reports.Insert(fmt.Sprintf("add->%d", port))
	}
	if reports.Len() == 0 {
		return nil
	}

	if !reflect.DeepEqual(updateService.Spec.Ports, currentService.Spec.Ports) {
		err = f.Client.Update(context.Background(), updateService, &client.UpdateOptions{})
		if err != nil {
			return err
		}
		for _, s := range reports.List() {
			f.Logger.Info(s)
		}
	}

	return nil
}
func (f *FalconController) callbackFunc(obj any) {

	g, ok := obj.(*v1alpha33.Gateway)
	if !ok {
		f.Logger.Error(errors.New("object type is not Gateway"), "type", obj)
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(g)
	if err != nil {
		rt.HandleError(err)
		return
	}
	f.Queue.Add(key)

}
func (f *FalconController) IsIgnoreGateway(gw *v1alpha33.Gateway) bool {
	v, found := gw.ObjectMeta.Labels[f.Options.IgnoreLabel]
	if found {
		return v == "true"
	}
	return false
}
func (f *FalconController) IsEffectivePort(port uint32) bool {
	if port <= 0 || port > 65535 {
		return false
	}
	return !lo.Contains[int32](f.Options.DefaultExposePorts, cast.ToInt32(port))
}
