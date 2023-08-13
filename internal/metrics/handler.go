package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sealbro/go-feed-me/internal/metrics/graphql"
	"net/http"
)

func HttpHandler() http.Handler {
	graphql.RegisterOnDefault()

	return promhttp.Handler()
}
