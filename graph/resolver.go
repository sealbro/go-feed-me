package graph

import (
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/notifier"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*storage.ArticleRepository
	*storage.ResourceRepository
	*notifier.SubscriptionManager[*model.FeedArticle]
	TracerProvider traces.ShutdownTracerProvider
}
