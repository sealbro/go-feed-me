package job

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"time"
)

type ParserFeedJob struct {
	feedParser         *gofeed.Parser
	logger             *logger.Logger
	articleRepository  *storage.ArticleRepository
	resourceRepository *storage.ResourceRepository
	manager            *notifier.SubscriptionManager[*model.FeedArticle]
	tracer             trace.Tracer
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
		tracer:             tracerProvider.Tracer("feed-parser-job"),
	}
}

func (p *ParserFeedJob) Execute(ctx context.Context) {
	ctx, span := p.tracer.Start(ctx, "execute")
	defer span.End()

	resources, err := p.resourceRepository.List(ctx, true)
	if err != nil || len(resources) == 0 {
		p.logger.Ctx(ctx).Warn("not found active resources")
		return
	}

	for _, resource := range resources {
		time.Sleep(3 * time.Second)

		if !p.processResource(ctx, resource) {
			return
		}
	}
}

func (p *ParserFeedJob) processResource(ctx context.Context, resource *storage.Resource) bool {
	ctx, span := p.tracer.Start(ctx, resource.Url)
	defer span.End()

	updatedResource, articles, err := p.fromUrl(ctx, *resource)
	if err != nil {
		p.logger.Ctx(ctx).Error("can't parse resource", zap.String("url", resource.Url))
		return true
	}

	if len(articles) == 0 {
		return true
	}

	p.notify(articles, resource)

	for _, article := range articles {
		err = p.articleRepository.Upsert(ctx, &article)
		if err != nil {
			p.logger.Ctx(ctx).Error("can't save article", zap.String("url", article.Link))
			return false
		} else {
			p.logger.Ctx(ctx).Info("article saved", zap.String("url", article.Link))
		}
	}

	err = p.resourceRepository.Upsert(ctx, updatedResource)
	if err != nil {
		p.logger.Ctx(ctx).Error("can't save resource", zap.String("url", resource.Url))
		return false
	} else {
		p.logger.Ctx(ctx).Info("resource saved", zap.String("url", resource.Url))
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
