package api

import (
	"net/http"

	"github.com/iyashjayesh/monigo/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(exporters.NewMonigoCollector())
}

func GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// PrometheusMetricsHandler handles the /metrics endpoint.
func PrometheusMetricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
