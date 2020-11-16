package contactcache

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

const (
	metricsNS = "contactcache"
)

var (
	metricsRequests = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNS,
		Name:      "requests",
		Help:      "histogram of requests through the middleware",
		Buckets:   []float64{},
	})

	cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNS,
		Name:      "cache_requests",
		Help:      "Cache hits and misses",
	}, []string{"type", "entity"})
)

func (s *Server) startMetricsEndpoint() error {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	s.log.Info("Starting metrics endpoint")

	metricsAddr := viper.GetString("metrics.address")
	if metricsAddr == "" {
		metricsAddr = ":9102"
	}

	return http.ListenAndServe(metricsAddr, r)
}
