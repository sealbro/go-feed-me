package job

import (
	"context"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/metrics"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"time"
)

type ParserFeedJob struct {
	feedParser         *gofeed.Parser
	logger             *logger.Logger
	articleRepository  *storage.ArticleRepository
	resourceRepository *storage.ResourceRepository
	manager            *notifier.SubscriptionManager[*model.FeedArticle]
	tracerProvider     traces.ShutdownTracerProvider
}

func NewParserFeedJob(logger *logger.Logger,
	articleRepository *storage.ArticleRepository,
	resourceRepository *storage.ResourceRepository,
	tracerProvider traces.ShutdownTracerProvider,
	manager *notifier.SubscriptionManager[*model.FeedArticle],
) quartz.Job {
	return &ParserFeedJob{
		logger:             logger,
		manager:            manager,
		feedParser:         gofeed.NewParser(),
		articleRepository:  articleRepository,
		resourceRepository: resourceRepository,
		tracerProvider:     tracerProvider,
	}
}

func (p *ParserFeedJob) Execute(ctx context.Context) error {
	tracer := p.tracerProvider.Tracer("feed-parser-job")
	ctx, span := tracer.Start(ctx, "execute")
	defer span.End()

	resources, err := p.resourceRepository.List(ctx, true)
	if err != nil || len(resources) == 0 {
		p.logger.WarnContext(ctx, "not found active resources")
		return err
	}

	span.AddEvent("resources", trace.WithAttributes(attribute.Key("resources.count").Int(len(resources))))

	for _, resource := range resources {
		time.Sleep(3 * time.Second)

		if !p.processResource(ctx, tracer, resource) {
			return fmt.Errorf("can't process resource: %s", resource.Url)
		}
	}

	return nil
}

func (p *ParserFeedJob) processResource(ctx context.Context, tracer trace.Tracer, resource *storage.Resource) bool {
	ctx, span := tracer.Start(ctx, resource.Url)
	defer span.End()

	updatedResource, articles, err := p.fromUrl(ctx, *resource)
	if err != nil {
		p.logger.WarnContext(ctx, "can't parse resource", slog.String("url", resource.Url))
		return true
	}

	if len(articles) == 0 {
		return true
	}

	// TODO: notify only new articles
	p.notify(articles, updatedResource)

	for _, article := range articles {
		err = p.articleRepository.Upsert(ctx, &article)
		if err != nil {
			p.logger.ErrorContext(ctx, "can't save article", slog.String("url", article.Link))
			return false
		} else {
			metrics.AddedArticlesCounter.Inc()
			p.logger.InfoContext(ctx, "article saved", slog.String("url", article.Link))
		}
	}

	err = p.resourceRepository.Upsert(ctx, updatedResource)
	if err != nil {
		p.logger.ErrorContext(ctx, "can't save resource", slog.String("url", resource.Url))
		return false
	} else {
		p.logger.InfoContext(ctx, "resource saved", slog.String("url", resource.Url))
	}

	return true
}

func (p *ParserFeedJob) notify(articles []storage.Article, resource *storage.Resource) {
	feedArticles := make([]*model.FeedArticle, len(articles))
	for i, article := range articles {
		feedArticles[i] = &model.FeedArticle{
			Created:       article.Created,
			Published:     article.Published,
			ResourceID:    article.ResourceId,
			ResourceTitle: resource.Title,
			Link:          article.Link,
			Title:         article.Title,
			Description:   article.Description,
			Content:       article.Content,
			Author:        article.Author,
			Image:         article.Image,
		}
	}
	p.manager.Notify(feedArticles...)
}

func (p *ParserFeedJob) Description() string {
	return "Feed parser"
}

func (p *ParserFeedJob) Key() int {
	return FeedParser
}

func (p *ParserFeedJob) fromUrl(ctx context.Context, resource storage.Resource) (*storage.Resource, []storage.Article, error) {
	url := resource.Url

	feed, err := p.feedParser.ParseURLWithContext(url, ctx)
	if err != nil || feed == nil || len(feed.Items) == 0 {
		return nil, nil, err
	}

	dateTimeNow := time.Now()

	var articles []storage.Article
	var maxPublished time.Time

	for _, item := range feed.Items {
		if item.PublishedParsed != nil && resource.Published.After(*item.PublishedParsed) {
			continue
		}

		published := dateTimeNow
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		author := ""
		if len(item.Authors) > 0 {
			author = item.Authors[0].Name
		}

		image := ""
		if item.Image != nil {
			image = item.Image.URL
		}

		articles = append(articles, storage.Article{
			ResourceId:  url,
			Created:     dateTimeNow,
			Link:        item.Link,
			Title:       item.Title,
			Description: item.Description,
			Content:     item.Content,
			Author:      author,
			Image:       image,
			Published:   published,
		})

		if published.After(maxPublished) {
			maxPublished = published
		}
	}

	return &storage.Resource{
		Modified:  dateTimeNow,
		Published: maxPublished.Add(time.Second),
		Url:       url,
		Title:     feed.Title,
		Active:    resource.Active,
	}, articles, nil
}
