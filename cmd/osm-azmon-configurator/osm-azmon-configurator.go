package main

import (
	"flag"
	"os"

	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/signals"
	"github.com/spf13/pflag"
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/draychev/osm-azmon-configurator/pkg/azmon/configurator"
	"github.com/draychev/osm-azmon-configurator/pkg/namespaces"
	"github.com/draychev/osm-azmon-configurator/pkg/version"
)

var (
	verbosity          string
	meshName           string // An ID that uniquely identifies an OSM instance
	kubeConfigFile     string
	osmNamespace       string
	azMonConfigMapName string

	scheme = runtime.NewScheme()

	flags = pflag.NewFlagSet(`osm-controller`, pflag.ExitOnError)
	log   = logger.New("osm-azmon-configurator/main")
)

func init() {
	flags.StringVarP(&verbosity, "verbosity", "v", "info", "Set log verbosity level")
	flags.StringVar(&meshName, "mesh-name", "", "OSM mesh name")
	flags.StringVar(&kubeConfigFile, "kubeconfig", "", "Path to Kubernetes config file.")
	flags.StringVar(&osmNamespace, "osm-namespace", "", "Namespace to which OSM belongs to.")
	flags.StringVar(&azMonConfigMapName, "azmon-configmap-name", "azmon-config", "Name of the Azure Monitor ConfigMap")

	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
}

func main() {
	log.Info().Msgf("Starting osm-azmon-configurator %s; %s; %s", version.Version, version.GitCommit, version.BuildDate)
	log.Info().Msgf("Log verbosity level set to %s", verbosity)
	log.Info().Msgf("Mesh Name is %s", meshName)
	log.Info().Msgf("OSM Namespace is %s", osmNamespace)
	log.Info().Msgf("Azure Monitor ConfigMap name is %s", azMonConfigMapName)

	if err := parseFlags(); err != nil {
		log.Fatal().Err(err).Msg("Error parsing cmd line arguments")
	}

	if err := logger.SetLogLevel(verbosity); err != nil {
		log.Fatal().Err(err).Msg("Error setting log level")
	}

	// Initialize kube config and client
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error creating kube config (kubeconfig=%s)", kubeConfigFile)
	}
	kubeClient := kubernetes.NewForConfigOrDie(kubeConfig)
	stop := signals.RegisterExitHandlers()

	nsMonitor, err := namespaces.NewNamespacesMonitor(kubeClient, meshName, stop)
	if err != nil {
		log.Err(err).Msgf("Error staring NamespacesMonitor.")
		return
	}

	if err := configurator.NewConfigurator(nsMonitor, kubeClient, osmNamespace, azMonConfigMapName, stop); err != nil {
		log.Err(err).Msgf("Error starting OSM AzMon Configurator")
		return
	}

	<-stop

	log.Info().Msgf("Stopping osm-controller %s; %s; %s", version.Version, version.GitCommit, version.BuildDate)
}

func parseFlags() error {
	if err := flags.Parse(os.Args); err != nil {
		return err
	}
	_ = flag.CommandLine.Parse([]string{})
	return nil
}
