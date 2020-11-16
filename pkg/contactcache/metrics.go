package contactcache

import (
	"fmt"
	"net/http"
	"time"

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
	metricsRequests = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricsNS,
		Name:      "requests",
		Help:      "histogram of requests through the middleware",
		Buckets: []float64{
			0.01, 0.1, 1, 10, 100, 1000, 2000, 5000,
		},
	}, []string{"status"})

	cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNS,
		Name:      "cache_requests",
		Help:      "Cache hits and misses",
	}, []string{"type", "entity"})
)

//startMetricsEndpoint starts a prometheus endpoint
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

//statusRecorder simple struct to record status from requests
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

//Metrics simple counter for requests
func (s *Server) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wRec := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		sTime := time.Now()
		next.ServeHTTP(wRec, r)
		metricsRequests.WithLabelValues(fmt.Sprintf("%d", wRec.Status)).Observe(float64(time.Since(sTime).Milliseconds()))
	})
}
