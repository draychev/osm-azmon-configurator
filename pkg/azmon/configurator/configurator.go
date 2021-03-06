package configurator

import (
	"context"
	"strings"
	"time"

	"github.com/openservicemesh/osm/pkg/logger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/draychev/osm-azmon-configurator/pkg/namespaces"
)

var (
	log = logger.New("osm-azmon-configurator/main")

	checkInterval = 1 * time.Minute
)

const configMapKey = "namespaces"

func NewConfigurator(nsMonitor *namespaces.NamespacesMonitor, kubeClient kubernetes.Interface, osmNamespace, azMonConfigMapName string, stop <-chan struct{}) error {

	makeConfigMap := func(namespacesJSON string) *v1.ConfigMap {
		log.Debug().Msgf("Adding namespaces %s", namespacesJSON)
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: osmNamespace,
				Name:      azMonConfigMapName,
			},
			Data: map[string]string{configMapKey: namespacesJSON},
		}
		return configMap
	}

	get := func() error {
		if _, err := kubeClient.CoreV1().ConfigMaps(osmNamespace).Get(context.Background(), azMonConfigMapName, metav1.GetOptions{}); err != nil {
			return err
		}
		return nil
	}

	create := func(namespacesJSON string) {

		if _, err := kubeClient.CoreV1().ConfigMaps(osmNamespace).Create(context.Background(), makeConfigMap(namespacesJSON), metav1.CreateOptions{}); err != nil {
			log.Err(err).Msgf("Error creating ConfigMap %s/%s", osmNamespace, azMonConfigMapName)
			return
		}
		log.Info().Msgf("Successfully created ConfigMap %s/%s", osmNamespace, azMonConfigMapName)
	}

	update := func(namespacesJSON string) {
		if _, err := kubeClient.CoreV1().ConfigMaps(osmNamespace).Update(context.Background(), makeConfigMap(namespacesJSON), metav1.UpdateOptions{}); err != nil {
			log.Err(err).Msgf("Error updating ConfigMap %s/%s", osmNamespace, azMonConfigMapName)
		}
	}

	refreshAzMonConfigMap := func() {
		log.Debug().Msgf("Refreshing %s/%s ConfigMap", osmNamespace, azMonConfigMapName)
		namespaceList, err := nsMonitor.ListMonitoredNamespaces()
		if err != nil {
			log.Err(err).Msg("Error listing namespaces")
			return
		}

		namespacesCSV := strings.Join(namespaceList, ",")

		if err := get(); err != nil {
			if errors.IsNotFound(err) {
				create(namespacesCSV)
			}
			log.Err(err).Msgf("Error looking up ConfigMap %s/%s", osmNamespace, azMonConfigMapName)
		} else {
			update(namespacesCSV)
		}
	}

	ticker := time.NewTicker(checkInterval)
	go func() {
		select {
		case <-nsMonitor.Events:
			log.Info().Msg("Triggered by Kubernetes Namespaces Event")
			refreshAzMonConfigMap()
		case <-ticker.C:
			log.Info().Msg("Triggered by Woken by Ticker")
			refreshAzMonConfigMap()
		case <-stop:
			return
		}
	}()

	return nil
}
