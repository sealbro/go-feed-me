package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type PublicApi struct {
	*PublicApiConfig
	Router *mux.Router
}

func NewPublicApi(config *PublicApiConfig) *PublicApi {
	return &PublicApi{
		PublicApiConfig: config,
		Router:          mux.NewRouter(),
	}
}

func (api *PublicApi) Build() *http.Server {
	return &http.Server{
		Addr:    api.Address,
		Handler: api.Router,
	}
}

func (api *PublicApi) Addr() string {
	return prettyAddress(api.Address)
}

func (api *PublicApi) RegisterRoutesFunc(fn func(router *mux.Router)) {
	fn(api.Router)
}

func (api *PublicApi) Prefix(serverName string, path string) string {
	return fmt.Sprintf("/%s/%s%s", api.ApplicationSlug, serverName, path)
}
