package api

type PublicApiConfig struct {
	Address         string `envconfig:"PUBLIC_ADDRESS" default:"127.0.0.1:8080"`
	ApplicationSlug string `envconfig:"SLUG" default:"feed"`
}
