package ops

import "github.com/prometheus/client_golang/prometheus"

var (
	roundsBucket = []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	roundsSuccessHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "engine",
		Subsystem: "match",
		Name:      "rounds_success",
		Help:      "",
		Buckets:   roundsBucket,
	}, []string{})

	roundsFailedHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "engine",
		Subsystem: "match",
		Name:      "rounds_failure",
		Help:      "",
		Buckets:   roundsBucket,
	}, []string{})

	roundsDurationHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "engine",
		Subsystem: "match",
		Name:      "rounds_duration_seconds",
		Help:      "",
	}, []string{})
)

func init() {
	prometheus.MustRegister(roundsSuccessHist)
	prometheus.MustRegister(roundsFailedHist)
	prometheus.MustRegister(roundsDurationHist)
}
