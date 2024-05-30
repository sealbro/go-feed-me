package api

type PrivateApiConfig struct {
	Address string `envconfig:"PRIVATE_ADDRESS" default:":8081"`
}
