package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type Registrar interface {
	Addr() string
	Build() *http.Server
	RegisterRoutesFunc(func(router *mux.Router))
	Prefix(serverName string, path string) string
}

func prettyAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return fmt.Sprintf("localhost%s", address)
	}

	return address
}
