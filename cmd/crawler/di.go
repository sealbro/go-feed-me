package main

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/api"
	"github.com/sealbro/go-feed-me/internal/graphql_api"
	"github.com/sealbro/go-feed-me/internal/job"
	"github.com/sealbro/go-feed-me/internal/metrics"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/subscribers"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/db"
	"github.com/sealbro/go-feed-me/pkg/graceful"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
	"os"
)

type jobGroup struct {
	dig.In
	Jobs []quartz.Job `group:"jobs"`
}

type CrawlerSettings struct {
	LoggerConfig *logger.Config
	DbConfig     *db.Config
	*api.PublicApiConfig
	*api.PrivateApiConfig
	*subscribers.DiscordConfig
	TracesConfig *traces.Config
	*job.DaemonConfig
}

func newSettings() (
	*CrawlerSettings,
	*logger.Config,
	*db.Config,
	*api.PublicApiConfig,
	*api.PrivateApiConfig,
	*subscribers.DiscordConfig,
	*traces.Config,
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

func provideApp() (graceful.Application, error) {
	container := dig.New()

	provideOrPanic(container, newSettings)
	provideOrPanic(container, logger.NewLogger)
	provideOrPanic(container, traces.NewTraceProvider)
	provideOrPanic(container, graceful.NewShutdownCloser)
	provideOrPanic(container, func() prometheusclient.Registerer {
		return prometheusclient.DefaultRegisterer
	})

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

	var app graceful.Application
	err := container.Invoke(func(application graceful.Application) {
		app = application
	})

	return app, err
}

func newApplication(logger *logger.Logger,
	collection *graceful.ShutdownCloser,
	daemon *job.Daemon,
	discordSubscriber *subscribers.DiscordSubscriber,
	publicApi *api.PublicApi,
	privateApi *api.PrivateApi,
	graphqlServer *graphql_api.GraphqlServer,
	tracerProvider traces.ShutdownTracerProvider,
	prometheusRegisterer prometheusclient.Registerer,
) graceful.Application {
	// Register prometheus metrics for promhttp.Handler()
	graphql_api.RegisterOn(prometheusRegisterer)
	metrics.RegisterOn(prometheusRegisterer)

	// Register and build api servers
	graphqlServer.RegisterRoutes(publicApi)
	privateApi.RegisterPrivateRoutes()
	publicServer := publicApi.Build()
	privateServer := privateApi.Build()

	// Setup tracer
	tracer := tracerProvider.Tracer("application")
	tracerCtx, span := tracer.Start(context.Background(), "graceful")

	return &graceful.Graceful{
		MainCtx: tracerCtx,
		Logger:  logger,
		StartAction: func(ctx context.Context) error {
			group, errCtx := errgroup.WithContext(ctx)

			group.Go(func() error {
				return daemon.Start(errCtx)
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
			defer func() {
				metrics.UnRegisterFrom(prometheusRegisterer)
				graphql_api.UnRegisterFrom(prometheusRegisterer)
			}()

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
		logger.Error("DI container registration wrong or does not exist", err)
		os.Exit(1)
	})
}
