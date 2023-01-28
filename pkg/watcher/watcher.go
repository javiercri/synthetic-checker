// ingresswatcher is a kubernetes controller that watches Ingress resources
// and configures several checks for each observed ingress.
package watcher

import (
	"fmt"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/luisdavim/synthetic-checker/pkg/checker"
	"github.com/luisdavim/synthetic-checker/pkg/watcher/filter"
	"github.com/luisdavim/synthetic-checker/pkg/watcher/ingress"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

type Options struct {
	Namespaces           string
	RequiredLabel        string
	MetricsAddr          string
	ProbeAddr            string
	EnableLeaderElection bool
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func StartBackground(chkr *checker.Runner, cfg Options) {
	go func() {
		err := Start(chkr, cfg)
		if err != nil {
			panic(err)
		}
	}()
}

func Start(chkr *checker.Runner, cfg Options) error {
	opts := zap.Options{
		Development: false,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrlOpts := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     cfg.MetricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: cfg.ProbeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       "synthetic-checker-controller",
	}

	nss := strings.Split(cfg.Namespaces, ",")

	if len(nss) == 1 {
		ctrlOpts.Namespace = nss[0]
	}

	if len(nss) > 1 {
		ctrlOpts.NewCache = cache.MultiNamespacedCacheBuilder(nss)
	}

	var filters []predicate.Predicate
	if len(nss) > 0 {
		filters = append(filters, filter.ByNamespace(nss))
	}

	if cfg.RequiredLabel != "" {
		filters = append(filters, filter.ByLabel(cfg.RequiredLabel))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlOpts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return fmt.Errorf("unable to start manager: %w", err)
	}

	if err = (&ingress.Reconciler{
		RequiredLabel: cfg.RequiredLabel,
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Checker:       chkr,
	}).SetupWithManager(mgr, filters); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ingress")
		return fmt.Errorf("unable to create controller: %w", err)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	setupLog.Info("starting controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return fmt.Errorf("problem running manager: %w", err)
	}
	return nil
}
