package k8s

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	informerResyncPeriod = 1 * time.Hour
)

type InformerAction uint16

const (
	InformerActionAdd InformerAction = iota
	InformerActionUpdate
	InformerActionDelete
)

type Client interface {
	GetService(ctx context.Context, callback func(InformerAction, *v1.Service)) error
}

type KubernetesClient struct {
	clientset *kubernetes.Clientset
	logger    *zap.Logger
}

func New(logger *zap.Logger, kubeconfig string) (*KubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesClient{
		logger:    logger,
		clientset: clientset
	}, nil
}

func (c *KubernetesClient) SubscribeToInformerEvents(ctx context.Context, handler func(event InformerEvent), options ...informers.SharedInformerOption) error {
	factory := informers.NewSharedInformerFactoryWithOptions(c.clientset, informerResyncPeriod, options...)

	svc := factory.Core().V1().Services()
	informer := svc.Informer()

	go factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			event := InformerEvent{NewState: new.(runtime.Object), Action: InformerActionAdd}
			handler(event)
		},
package k8s

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	informerResyncPeriod = 1 * time.Hour
)

type InformerAction uint16

const (
	InformerActionAdd InformerAction = iota
	InformerActionUpdate
	InformerActionDelete
)

type Client interface {
	GetService(ctx context.Context, callback func(InformerAction, *v1.Service)) error
}

type KubernetesClient struct {
	clientset *kubernetes.Clientset
	logger    *zap.Logger
}

func New(logger *zap.Logger, kubeconfig string) (*KubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesClient{
		logger:    logger,
		clientset: clientset,
	}, nil
}

func (c *KubernetesClient) SubscribeToInformerEvents(ctx context.Context, handler func(event InformerEvent), options ...informers.SharedInformerOption) error {
	factory := informers.NewSharedInformerFactoryWithOptions(c.clientset, informerResyncPeriod, options...)

	svc := factory.Core().V1().Services()
	informer := svc.Informer()

	go factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("tempo limite excedido ao aguardar a sincronização dos caches")
	}

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			event := InformerEvent{NewState: new.(runtime.Object), Action: InformerActionAdd}
			
			handler(event)
		},
		
		UpdateFunc: func(old interface{}, new interface{}) {
			event := InformerEvent{NewState: new.(runtime.Object), OldState: old.(runtime.Object), Action: InformerActionUpdate}
			
			handler(event)
		},
		
		DeleteFunc: func(old interface{}) {
			event := InformerEvent{NewState: old.(runtime.Object), Action: InformerActionDelete}
			
			handler(event)
		}
	})

	return err
}
