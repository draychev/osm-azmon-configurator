package namespaces

import (
	"errors"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/logger"
)

var (
	log = logger.New("kube-controller")

	errSyncingInformerCache = errors.New("error syncing informer cache")
)

const (
	defaultKubeEventResyncInterval = 5 * time.Minute
)

// NamespacesMonitor is a struct for all components necessary to connect to and maintain state of a Kubernetes cluster.
type NamespacesMonitor struct {
	namespaceInformer cache.SharedIndexInformer
	Events            chan interface{}
}

// NewNamespacesMonitor returns a new kubernetes.Controller which means to provide access to locally-cached k8s resources
func NewNamespacesMonitor(kubeClient kubernetes.Interface, meshName string, stop <-chan struct{}) (*NamespacesMonitor, error) {
	monitorNamespaceLabel := map[string]string{constants.OSMKubeResourceMonitorAnnotation: meshName}
	labelSelector := fields.SelectorFromSet(monitorNamespaceLabel).String()
	option := informers.WithTweakListOptions(func(opt *metav1.ListOptions) {
		opt.LabelSelector = labelSelector
	})

	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultKubeEventResyncInterval, option)
	namespaceInformer := informerFactory.Core().V1().Namespaces().Informer()

	events := make(chan interface{})

	namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { events <- nil },
		UpdateFunc: func(oldObj, newObj interface{}) { events <- nil },
		DeleteFunc: func(obj interface{}) { events <- nil },
	})

	go namespaceInformer.Run(stop)

	log.Info().Msg("Namespace informer goroutine started. Waiting for Namespaces informer cache to sync.")

	if !cache.WaitForCacheSync(stop, namespaceInformer.HasSynced) {
		return nil, errSyncingInformerCache
	}

	log.Info().Msg("DONE: Namespaces informer cache synced.")

	return &NamespacesMonitor{
		namespaceInformer: namespaceInformer,
		Events:            events,
	}, nil
}

// ListMonitoredNamespaces returns all namespaces that the mesh is monitoring.
func (c NamespacesMonitor) ListMonitoredNamespaces() ([]string, error) {
	var namespaces []string
	for _, ns := range c.namespaceInformer.GetStore().List() {
		if namespace, ok := ns.(*corev1.Namespace); ok {
			namespaces = append(namespaces, namespace.Name)
		}
	}
	return namespaces, nil
}
