package graphql_api

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sealbro/go-feed-me/graph"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/api"
	"github.com/sealbro/go-feed-me/internal/metrics"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/internal/traces"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"log/slog"
	"net/http"
	"time"
)

type GraphqlServer struct {
	resolvers *graph.Resolver
	logger    *logger.Logger
}

func NewGraphqlServer(logger *logger.Logger,
	articleRepository *storage.ArticleRepository,
	resourceRepository *storage.ResourceRepository,
	tracerProvider traces.ShutdownTracerProvider,
	subscriptionManager *notifier.SubscriptionManager[*model.FeedArticle]) *GraphqlServer {
	graphqlApi := &GraphqlServer{
		resolvers: &graph.Resolver{
			ArticleRepository:   articleRepository,
			ResourceRepository:  resourceRepository,
			SubscriptionManager: subscriptionManager,
			TracerProvider:      tracerProvider,
		},
		logger: logger,
	}

	return graphqlApi
}

func (server *GraphqlServer) RegisterRoutes(registrar api.Registrar) {
	urlPrefix := "graphql"

	schema := graph.NewExecutableSchema(graph.Config{Resolvers: server.resolvers})
	srv := handler.NewDefaultServer(schema)

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})
	srv.Use(metrics.NewPrometheusMetricsExtension())
	srv.Use(traces.NewTraceExtension(server.resolvers.TracerProvider))

	endpoint := registrar.Prefix(urlPrefix, "/query")
	playgroundEndpoint := registrar.Prefix(urlPrefix, "/")

	registrar.RegisterRoutesFunc(func(router *mux.Router) {
		router.Handle(playgroundEndpoint, playground.Handler("GraphQL playground", endpoint))
		router.Handle(endpoint, srv)
	})

	graphqlUrl := fmt.Sprintf("http://%s%s", registrar.Addr(), playgroundEndpoint)

	server.logger.Info("GraphQl server", slog.String("url", graphqlUrl))
}
