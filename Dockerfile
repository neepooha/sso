# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.1
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /apps/sso

RUN --mount=type=cache,target=/go/pkg/mod/ \
--mount=type=bind,source=go.sum,target=go.sum \
--mount=type=bind,source=go.mod,target=go.mod \
go mod download -x

FROM base AS build-sso
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    go build -o /sso ./cmd/sso/main.go

FROM scratch AS server
COPY ./config.env ./
COPY ./config ./config
COPY ./migrations ./migrations
COPY --from=build-sso /sso ./
EXPOSE 44044 
ENTRYPOINT [ "./sso" ]