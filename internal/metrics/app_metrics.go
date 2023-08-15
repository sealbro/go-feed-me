package metrics

import prometheusclient "github.com/prometheus/client_golang/prometheus"

var (
	AddedArticlesCounter  prometheusclient.Counter
	AddedResourcesCounter prometheusclient.Counter
)

func RegisterOn(registerer prometheusclient.Registerer) {
	AddedArticlesCounter = prometheusclient.NewCounter(
		prometheusclient.CounterOpts{
			Name: "feed_articles_total",
			Help: "Total number of added articles started on the graphql server.",
		},
	)

	AddedResourcesCounter = prometheusclient.NewCounter(
		prometheusclient.CounterOpts{
			Name: "feed_resources_total",
			Help: "Total number of added resources started on the graphql server.",
		},
	)

	registerer.MustRegister(
		AddedArticlesCounter,
		AddedResourcesCounter,
	)
}

func UnRegisterFrom(registerer prometheusclient.Registerer) {
	registerer.Unregister(AddedArticlesCounter)
	registerer.Unregister(AddedResourcesCounter)
}
