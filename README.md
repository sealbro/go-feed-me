# RSS feed parser for go

## Features

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

### Mutations

```graphql
mutation AddResources {
  addResources (resources: [
    {url: "https://github.com/opencv/opencv/releases.atom"},
    {url:"https://github.com/openvinotoolkit/openvino/releases.atom"},
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