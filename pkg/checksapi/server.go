package checksapi

import (
	"time"

	"github.com/luisdavim/synthetic-checker/pkg/checker"
	"github.com/luisdavim/synthetic-checker/pkg/server"
)

type Options struct {
	FailStatus     int
	DegradedStatus int
	ExposeConfig   bool
}

// New creates a new check API server
func New(chkr *checker.Runner, srvCfg server.Config, opts Options) *server.Server {
	srv := server.New(srvCfg)
	srv.WithShutdownFunc(func() error {
		// ensure the checker routines are stopped
		chkr.Stop()
		time.Sleep(2 * time.Second)
		return nil
	})
	setRoutes(chkr, srv, opts)

	return srv
}
