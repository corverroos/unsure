package unsure

import (
	"flag"
	"log/syslog"
	"net"
	"net/http"
	"os"
	"time"

	klog "github.com/go-kit/kit/log"
	ksyslog "github.com/go-kit/kit/log/syslog"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	pushhandler "github.com/prometheus/pushgateway/handler"
	"github.com/prometheus/pushgateway/storage"
)

var (
	promJob         = flag.String("prom_job", "unsure", "prometheus metrics job name")
	promPushAddress = flag.String("prom_push_address", "localhost:9091", "prometheus pushgateway address")
	promPush        = flag.Bool("prom_push", true, "enable pushing prometheus metrics on shutdown")

	reflexSoftErrorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "unsure",
		Subsystem: "reflex",
		Name:      "soft_errored_total",
		Help:      "Total number of times a reflex consumer returned a soft error",
	}, []string{"consumer"})

	reflexHardErrorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "unsure",
		Subsystem: "reflex",
		Name:      "hard_errored_total",
		Help:      "Total number of times a reflex consumer returned a hard error",
	}, []string{"consumer"})
)

func init() {
	prometheus.MustRegister(reflexSoftErrorCounter)
	prometheus.MustRegister(reflexHardErrorCounter)
}

func pushMetrics() error {
	if !*promPush {
		return nil
	}
	return push.New(*promPushAddress, *promJob).Gatherer(prometheus.DefaultGatherer).Push()
}

func StartPromServer() (storage.MetricStore, error) {
	*promPush = false // If serving, don't push

	const storePath = "/tmp/arena/prometheus.store"
	if err := os.RemoveAll(storePath); err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", *promPushAddress)
	if err != nil {
		return nil, err
	}

	logger, err := newKLogger()
	if err != nil {
		return nil, err
	}

	ms := storage.NewDiskMetricStore(storePath, time.Second, nil, logger)

	handler := pushhandler.Push(ms, true, false, logger)
	wrapHandler := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p)
	}

	router := httprouter.New()
	router.PUT("/metrics/job/:job/*labels", wrapHandler)
	router.PUT("/metrics/job/:job", wrapHandler)

	go func() {
		Fatal(http.Serve(l, router))
	}()

	return ms, nil
}

// newKLogger returns a syslog go-kit logger.
func newKLogger() (klog.Logger, error) {
	w, err := syslog.New(syslog.LOG_INFO, "experiment")
	if err != nil {
		return nil, err
	}

	return ksyslog.NewSyslogLogger(w, klog.NewLogfmtLogger), nil
}
