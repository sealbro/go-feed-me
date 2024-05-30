package subscribers

import (
	"context"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/pkg/graceful"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
)

type DiscordSubscriber struct {
	subscriptionManager *notifier.SubscriptionManager[*model.FeedArticle]
	cancelFunc          context.CancelFunc
	config              *DiscordConfig
	logger              *logger.Logger
	converter           *md.Converter
}

func NewDiscordSubscriber(logger *logger.Logger, config *DiscordConfig,
	subscriptionManager *notifier.SubscriptionManager[*model.FeedArticle], closer *graceful.ShutdownCloser) *DiscordSubscriber {
	d := &DiscordSubscriber{
		logger:              logger,
		subscriptionManager: subscriptionManager,
		config:              config,
		converter:           md.NewConverter("", true, nil),
	}

	closer.Register(d)

	return d
}

func (s *DiscordSubscriber) Subscribe(ctx context.Context) error {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	s.cancelFunc = cancelFunc
	events, err := s.subscriptionManager.AddSubscriber(cancelCtx, "discord")
	if err != nil {
		return err
	}

	go s.processEvents(events)

	return nil
}

func (s *DiscordSubscriber) processEvents(fireEvents <-chan []*model.FeedArticle) {
	client := webhook.New(snowflake.ID(s.config.WebhookId), s.config.WebhookToken)

	for events := range fireEvents {

		embeds := make([]discord.Embed, len(events))
		for i, event := range events {
			revertIndex := len(events) - 1 - i
			embeds[revertIndex] = discord.Embed{
				Title:       event.Title,
				Type:        discord.EmbedTypeRich,
				Description: event.Description,
				URL:         event.Link,
				Timestamp:   &event.Published,
				Color:       0x87CEEB,
				Footer: &discord.EmbedFooter{
					Text: event.ResourceTitle,
				},
				Author: &discord.EmbedAuthor{Name: event.Author},
			}
		}

		_, err := client.CreateEmbeds(embeds)
		if err != nil {
			s.logger.Error("Failed to send message to discord", err)
		}
	}
}

func (s *DiscordSubscriber) Close() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
	return nil
}
