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
	"go.uber.org/zap"
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
	cancel, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	events, err := s.subscriptionManager.AddSubscriber(cancel)
	if err != nil {
		return err
	}

	s.cancelFunc = cancelFunc

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
				Type:        discord.EmbedTypeArticle,
				Description: event.Description,
				URL:         event.Link,
				Timestamp:   &event.Published,
				Color:       0x87CEEB,
				Footer: &discord.EmbedFooter{
					Text: event.ResourceTitle,
				},
				Author: &discord.EmbedAuthor{Name: event.Author},
			}

			//convertString, err := s.converter.ConvertString(event.Content)
			//if err != nil {
			//	s.logger.Error("Failed to convert html to markdown", zap.Error(err))
			//	continue
			//}
			//if len(convertString) > 2000 {
			//	convertString = convertString[:1980] + "... :scissors:"
			//}
			//content, err := client.CreateContent(convertString)
			//if err != nil {
			//	s.logger.Error("Failed to send message to discord", zap.Error(err))
			//} else {
			//	s.logger.Info("Message sent to discord", zap.String("message_id", content.ID.String()))
			//}
		}

		_, err := client.CreateEmbeds(embeds)
		if err != nil {
			s.logger.Error("Failed to send message to discord", zap.Error(err))
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
