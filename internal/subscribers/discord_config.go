package subscribers

type DiscordConfig struct {
	WebhookId    uint64 `envconfig:"DISCORD_WEBHOOK_ID"`
	WebhookToken string `envconfig:"DISCORD_WEBHOOK_TOKEN"`
}
