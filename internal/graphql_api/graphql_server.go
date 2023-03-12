package graphql_api

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/sealbro/go-feed-me/graph"
	"github.com/sealbro/go-feed-me/graph/model"
	"github.com/sealbro/go-feed-me/internal/api"
	"github.com/sealbro/go-feed-me/internal/storage"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type GraphqlServer struct {
	*api.PublicApi

	resolvers *graph.Resolver
	logger    *logger.Logger
}

func NewGraphqlServer(logger *logger.Logger,
	api *api.PublicApi,
	articleRepository *storage.ArticleRepository,
	resourceRepository *storage.ResourceRepository,
	subscriptionManager *notifier.SubscriptionManager[*model.FeedArticle]) *GraphqlServer {
	graphqlApi := &GraphqlServer{
		resolvers: &graph.Resolver{
			ArticleRepository:   articleRepository,
			ResourceRepository:  resourceRepository,
			SubscriptionManager: subscriptionManager,
		},
		PublicApi: api,
		logger:    logger,
	}

	return graphqlApi
}

func (server *GraphqlServer) RegisterRoutes() {
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

	endpoint := server.Prefix(urlPrefix, "/query")
	playgroundEndpoint := server.Prefix(urlPrefix, "/")
	server.Router.Handle(playgroundEndpoint, playground.Handler("GraphQL playground", endpoint))
	server.Router.Handle(endpoint, srv)

	graphqlUrl := fmt.Sprintf("http://%s%s", server.PublicApi.Address, playgroundEndpoint)

	server.logger.Info("GraphQl server", zap.String("url", graphqlUrl))
}
