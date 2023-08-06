package api

type PrivateApiConfig struct {
	Address string `envconfig:"PRIVATE_ADDRESS" default:"127.0.0.1:8081"`
}
