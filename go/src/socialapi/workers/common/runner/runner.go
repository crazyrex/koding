package runner

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"socialapi/config"
	"socialapi/workers/helper"
	"syscall"

	"github.com/koding/bongo"
	"github.com/koding/logging"
	"github.com/koding/worker"
)

var (
	flagConfFile = flag.String("c", "", "Configuration profile from file")
	flagDebug    = flag.Bool("d", false, "Debug mode")
	flagVersion  = flag.Int("v", 0, "Worker Version")
)

type Runner struct {
	Log      logging.Logger
	Conf     *config.Config
	Bongo    *bongo.Bongo
	Listener *worker.Listener
	Name     string
}

func New(name string) *Runner {
	return &Runner{Name: name}
}

func WrapWithVersion(name string, version *int) string {
	return fmt.Sprintf("%s:%d", name, *version)
}

func (r *Runner) Init() error {
	flag.Parse()
	if *flagConfFile == "" {
		return fmt.Errorf("Please define config file with -c Exiting...")
	}

	r.Conf = config.MustRead(*flagConfFile)

	// create logger for our package
	r.Log = helper.CreateLogger(
		WrapWithVersion(r.Name, flagVersion),
		*flagDebug,
	)

	// panics if not successful
	r.Bongo = helper.MustInitBongo(
		WrapWithVersion(r.Name, flagVersion),
		WrapWithVersion(r.Conf.EventExchangeName, flagVersion),
		r.Conf,
		r.Log,
	)

	return nil
}

func (r *Runner) Listen(handler worker.Handler) {
	r.Listener = worker.NewListener(
		WrapWithVersion(r.Name, flagVersion),
		WrapWithVersion(r.Conf.EventExchangeName, flagVersion),
		r.Log,
	)

	// blocking
	// listen for events
	r.Listener.Listen(helper.NewRabbitMQ(r.Conf, r.Log), handler)
}

func (r *Runner) Close() {
	if r.Listener != nil {
		r.Listener.Close()
	}
	r.Bongo.Close()
}

func (r *Runner) RegisterSignalHandler(f func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	for {
		signal := <-signals
		switch signal {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP:
			f()
			r.Close()
			os.Exit(1)
		}
	}
}
