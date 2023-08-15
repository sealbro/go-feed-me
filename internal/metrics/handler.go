package metrics

import (
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sealbro/go-feed-me/internal/metrics/graphql"
	"net/http"
)

func HttpHandler() http.Handler {
	registerer := prometheusclient.DefaultRegisterer
	graphql.RegisterOn(registerer)
	RegisterOn(registerer)

	return promhttp.Handler()
}

func NewPrometheusMetricsExtension() graphql.GraphqlPrometheusMetrics {
	return graphql.GraphqlPrometheusMetrics{}
}
