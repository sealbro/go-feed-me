package main

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/api"
	"github.com/sealbro/go-feed-me/internal/graphql_api"
	"github.com/sealbro/go-feed-me/internal/job"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/subscribers"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/db"
	"github.com/sealbro/go-feed-me/pkg/graceful"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type jobGroup struct {
	dig.In
	Jobs []quartz.Job `group:"jobs"`
}

type CrawlerSettings struct {
	*logger.LoggerConfig
	*db.DbConfig
	*api.PublicApiConfig
	*api.PrivateApiConfig
	*subscribers.DiscordConfig
	*traces.TracesConfig
	*job.DaemonConfig
}

var diContainer *dig.Container

func init() {
	container := dig.New()

	provideOrPanic(container, newSettings)
	provideOrPanic(container, logger.NewLogger)
	provideOrPanic(container, graceful.NewShutdownCloser)

	provideOrPanic(container, traces.NewTraceProvider)
	provideOrPanic(container, logger.NewGormLogger)
	provideOrPanic(container, db.NewDatabase)
	provideOrPanic(container, storage.NewResourceRepository)
	provideOrPanic(container, storage.NewArticleRepository)

	provideOrPanic(container, notifier.NewSubscriptionManager[*model.FeedArticle])
	provideOrPanic(container, subscribers.NewDiscordSubscriber)
	provideOrPanic(container, job.NewDaemon)
	provideOrPanic(container, job.NewParserFeedJob, dig.Group("jobs"))
	provideOrPanic(container, func(group jobGroup) []quartz.Job { return group.Jobs })

	provideOrPanic(container, api.NewPublicApi)
	provideOrPanic(container, api.NewPrivateApi)
	provideOrPanic(container, graphql_api.NewGraphqlServer)

	provideOrPanic(container, newApplication)

	diContainer = container
}

func newSettings() (
	*CrawlerSettings,
	*logger.LoggerConfig,
	*db.DbConfig,
	*api.PublicApiConfig,
	*api.PrivateApiConfig,
	*subscribers.DiscordConfig,
	*traces.TracesConfig,
	*job.DaemonConfig,
) {
	settings := &CrawlerSettings{}

	err := envconfig.Process("", settings)
	if err != nil {
		panic(fmt.Errorf("can not load settings: %v", err))
	}

	return settings,
		settings.LoggerConfig,
		settings.DbConfig,
		settings.PublicApiConfig,
		settings.PrivateApiConfig,
		settings.DiscordConfig,
		settings.TracesConfig,
		settings.DaemonConfig
}

func newApplication(logger *logger.Logger,
	collection *graceful.ShutdownCloser,
	daemon *job.Daemon,
	discordSubscriber *subscribers.DiscordSubscriber,
	publicApi *api.PublicApi,
	privateApi *api.PrivateApi,
	graphqlServer *graphql_api.GraphqlServer,
	tracerProvider traces.ShutdownTracerProvider,
) graceful.Application {

	graphqlServer.RegisterRoutes(publicApi)
	privateApi.RegisterPrivateRoutes()

	publicServer := publicApi.Build()
	privateServer := privateApi.Build()

	tracer := tracerProvider.Tracer("application")
	tracerCtx, span := tracer.Start(context.Background(), "graceful")

	return &graceful.Graceful{
		MainCtx: tracerCtx,
		Logger:  logger,
		StartAction: func(ctx context.Context) error {
			group, errCtx := errgroup.WithContext(ctx)

			group.Go(func() error {
				daemon.Start(errCtx)
				return nil
			})

			group.Go(func() error {
				return discordSubscriber.Subscribe(errCtx)
			})

			group.Go(func() error {
				return privateServer.ListenAndServe()
			})

			group.Go(func() error {
				return publicServer.ListenAndServe()
			})

			return group.Wait()
		},
		ShutdownAction: func(ctx context.Context) error {
			defer span.End()

			group, errCtx := errgroup.WithContext(ctx)

			group.Go(func() error {
				return collection.Close()
			})

			group.Go(func() error {
				return privateServer.Shutdown(errCtx)
			})

			group.Go(func() error {
				return publicServer.Shutdown(errCtx)
			})

			group.Go(func() error {
				return tracerProvider.Shutdown(errCtx)
			})

			return group.Wait()
		},
	}
}

func provideOrPanic(container *dig.Container, constructor interface{}, opts ...dig.ProvideOption) {
	err := container.Provide(constructor, opts...)
	if err == nil {
		return
	}

	_ = container.Invoke(func(logger *logger.Logger) {
		logger.Fatal("container.Provide", zap.Error(err))
	})
}
