FROM golang as builder
WORKDIR /build

COPY *.go go.mod go.sum config.yaml .
RUN go get ./...  && \
    CGO_ENABLED=0 go build -o gitea-webhook *.go

FROM alpine
WORKDIR /app

COPY --from=builder /build/gitea-webhook /build/config.yaml .
CMD ["./gitea-webhook"]