package api

import (
	"github.com/gorilla/mux"
)

type PrivateApi struct {
	Router *mux.Router
}

func NewPrivateApi() *PrivateApi {
	return &PrivateApi{Router: mux.NewRouter()}
}
