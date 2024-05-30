package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"log/slog"
	"net/http"
)

type PrivateApi struct {
	Router  *mux.Router
	Address string
	logger  *logger.Logger
}

func NewPrivateApi(logger *logger.Logger, config *PrivateApiConfig) *PrivateApi {
	return &PrivateApi{
		Address: config.Address,
		Router:  mux.NewRouter(),
		logger:  logger,
	}
}

func (a *PrivateApi) Build() *http.Server {
	return &http.Server{
		Addr:    a.Address,
		Handler: a.Router,
	}
}

func (a *PrivateApi) Addr() string {
	return prettyAddress(a.Address)
}

func (a *PrivateApi) RegisterRoutesFunc(fn func(router *mux.Router)) {
	fn(a.Router)
}

func (a *PrivateApi) Prefix(serverName string, path string) string {
	return fmt.Sprintf("/%s/%s", serverName, path)
}

func (a *PrivateApi) RegisterPrivateRoutes() {
	a.Router.HandleFunc("/liveness", a.liveness).Methods("GET")
	a.Router.HandleFunc("/readiness", a.readiness).Methods("GET")
	a.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	a.logger.Info("Private server", slog.String("url", fmt.Sprintf("http://%s%s", a.Addr(), "/metrics")))
}

func (a *PrivateApi) liveness(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("healthy"))
}

func (a *PrivateApi) readiness(writer http.ResponseWriter, _ *http.Request) {
	// TODO check db connection
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ready"))
}
