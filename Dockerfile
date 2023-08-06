FROM golang:1.20-bullseye as builder

WORKDIR /src
COPY . .

WORKDIR /src/cmd/crawler
RUN CGO_ENABLED=1 go build -o /bin/runner

FROM gcr.io/distroless/base as runtime

COPY --from=builder /bin/runner /

CMD ["/runner"]
