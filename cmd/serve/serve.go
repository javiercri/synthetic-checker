package serve

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/luisdavim/synthetic-checker/pkg/checker"
	"github.com/luisdavim/synthetic-checker/pkg/checksapi"
	"github.com/luisdavim/synthetic-checker/pkg/config"
	"github.com/luisdavim/synthetic-checker/pkg/leaderelection"
	"github.com/luisdavim/synthetic-checker/pkg/server"
	"github.com/luisdavim/synthetic-checker/pkg/watcher"
)

type options struct {
	haMode         bool
	watchIngresses bool
	leID           string
	leNs           string
	reset          bool
}

func New(cfg *config.Config) *cobra.Command {
	var (
		opts    options
		wOpts   watcher.Options
		apiOpts checksapi.Options
	)
	// cmd represents the base command when called without any subcommands
	cmd := &cobra.Command{
		Use:          "serve",
		Aliases:      []string{"run", "start"},
		Short:        "Run as a service",
		Long:         `Run as a service.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			chkr, err := checker.NewFromConfig(*cfg, !opts.haMode)
			if err != nil {
				return err
			}

			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Println("Config file changed:", e.Name)
				if err := viper.Unmarshal(cfg, config.DecodeHooks()); err != nil {
					panic(err)
				}
				if err := chkr.ReloadConfig(*cfg, !opts.haMode, opts.reset); err != nil {
					panic(err)
				}
			})
			viper.WatchConfig()

			var srvCfg server.Config
			if err := server.ReadConfig(&srvCfg); err != nil {
				return fmt.Errorf("error reading server config: %v", err)
			}

			wOpts.MetricsAddr = fmt.Sprintf(":%d", srvCfg.HTTP.Port+1)
			wOpts.ProbeAddr = fmt.Sprintf(":%d", srvCfg.HTTP.Port+2)

			if opts.haMode {
				le, err := leaderelection.NewLeaderElector(opts.leID, opts.leNs)
				if err != nil {
					return err
				}
				go le.RunLeaderElection(context.Background(),
					func(ctx context.Context) {
						chkr.Run(ctx)
						if opts.watchIngresses {
							watcher.StartBackground(chkr, wOpts)
						}
						<-ctx.Done() // hold the routine, Run goes into the background
					},
					chkr.StatusSyncer(le.ID, false, srvCfg.HTTP.Port),
					func() {
						chkr.Stop()
						os.Exit(1) // TODO: is this overkill?
					},
				)
			} else {
				chkr.Run(context.Background())
				if opts.watchIngresses {
					watcher.StartBackground(chkr, wOpts)
				}
			}

			srv := checksapi.New(chkr, srvCfg, apiOpts)
			srv.Run()
			return nil
		},
	}

	server.Init(cmd)

	cmd.Flags().BoolVarP(&opts.reset, "reset-config-on-reload", "", false, "delete all existing checks when hot-reloading the config file")
	cmd.Flags().BoolVarP(&opts.haMode, "k8s-leader-election", "", false, "Enable leader election, only works when running in k8s")
	cmd.Flags().StringVarP(&opts.leID, "leader-election-id", "", "", "set the leader election ID, defaults to the pod IP or hostname")
	cmd.Flags().StringVarP(&opts.leNs, "leader-election-ns", "", "", "set the leader election namespace, defaults to the current namespace")
	cmd.Flags().IntVarP(&apiOpts.FailStatus, "failed-status-code", "F", http.StatusOK, "HTTP status code to return when all checks are failed")
	cmd.Flags().IntVarP(&apiOpts.DegradedStatus, "degraded-status-code", "D", http.StatusOK, "HTTP status code to return when check check is failed")
	cmd.Flags().BoolVarP(&apiOpts.ExposeConfig, "expose-config", "", false, "Enable the /config endpoint to expose the runtime configuration")
	cmd.Flags().BoolVarP(&opts.watchIngresses, "watch-ingresses", "w", false, "Automatically setup checks for k8s ingresses, only works when running in k8s")
	cmd.Flags().StringVarP(&wOpts.RequiredLabel, "required-label", "", "", "ignore Ingress resources that don't have this label set to a truthful value")
	cmd.Flags().StringVarP(&wOpts.Namespaces, "namespaces", "n", "", "a comma separated list of namespaces from where to watch Ingresses")

	return cmd
}
