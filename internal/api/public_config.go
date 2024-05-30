package api

type PublicApiConfig struct {
	Address         string `envconfig:"PUBLIC_ADDRESS" default:":8080"`
	ApplicationSlug string `envconfig:"SLUG" default:"feed"`
}
