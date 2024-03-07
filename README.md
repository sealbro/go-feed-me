# RSS feed notifier as a service

[![Hub](https://badgen.net/docker/pulls/sealbro/go-feed-me?icon=docker&label=go-feed-me)](https://hub.docker.com/r/sealbro/go-feed-me/)

## Features

- [x] GraphQL UI and API
- [x] Store in database (sqlite, postgres)
- [x] Add new RSS feed resources and store them
- [x] Fetch new articles from RSS feed resources
- [x] Notify new articles to graphql subscribers, discord.
- [x] Observability (logs, metrics, traces)


## Quick start

```bash
docker run -it --rm -p 8080:8080 -p 8081:8081 sealbro/go-feed-me:latest
```

### Environment variables

| Name                            | Description                | Default          |
|---------------------------------|----------------------------|------------------|
| `SLUG`                          | Path prefix                | `feed`           |
| `CRON`                          | Cron pattern when run jobs | `1/60 * * * * *` |
| `SQLITE_CONNECTION`             | Sqlite file location       | `feed.db`        |
| `POSTGRES_CONNECTION`           | Postgres connection string | empty            |
| `DISCORD_WEBHOOK_ID`            | Discord webhook id         | empty            |
| `DISCORD_WEBHOOK_TOKEN`         | Discord webhook token      | empty            |
| `LOG_LEVEL`                     | slog level                 | `INFO`           |
| `OTEL_EXPORTER_OTLP_ENDPOINT`   | Otlp grpc endpoint         | empty            |

- Postgres's [connection string](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL): `host=<ip or host> user=<username> password=<password> dbname=feed port=5432 sslmode=disable`
- Discord how get id and token for [webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
- Cron pattern [quartz](https://github.com/reugn/go-quartz)

## Graphql

### Queries

```graphql
query Articles {
    articles (after: "2023-01-01T15:04:05.999999999Z") {
        published,
        link,
        title,
        description,
        content
    }
}
```

```graphql
query Resources {
    resources(active: true) {
        url
        title
        active
        created
        modified
        published
    }
}
```

### Mutations

```graphql
mutation AddResources {
    addResources (resources: [
        {url: "https://github.com/opencv/opencv/releases.atom"},
        {url: "https://github.com/openvinotoolkit/openvino/releases.atom"},
        {url: "https://github.com/hybridgroup/gocv/releases.atom"},
    ]) 
}
```

### Subscriptions

```graphql
subscription notifyNewData {
    articles {
        title
        description
        content
        link
    }
}
```