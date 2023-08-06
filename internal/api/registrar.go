package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Registrar interface {
	Addr() string
	Build() *http.Server
	RegisterRoutesFunc(func(router *mux.Router))
	Prefix(serverName string, path string) string
}
