FROM golang:1.14-alpine3.11 as build

LABEL maintainer="https://github.com/yosuke0517"

WORKDIR /go/app

COPY . .

ENV GO111MODULE=off

RUN set -eux && \
  apk update && \
  apk add --no-cache git curl && \
  curl -fLo /go/bin/air https://git.io/linux_air && \
  chmod +x /go/bin/air && \
  go get -u github.com/labstack/echo/... && \
  go get -u github.com/go-delve/delve/cmd/dlv && \
  go build -o /go/bin/dlv github.com/go-delve/delve/cmd/dlv

ENV GO111MODULE on

RUN set -eux && \
  go build -o system-trade-api ./main.go

FROM alpine:3.11

WORKDIR /app

COPY --from=build /go/app/system-trade-api .

RUN set -x && \
  addgroup go && \
  adduser -D -G go go && \
  chown -R go:go /app/system-trade-api

CMD ["./system-trade-api"]