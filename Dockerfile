# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.1
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /apps/sso

COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base as build-migrate
CMD [ "go run ./cmd/migrator/main.go" ] 

FROM base as build-sso
RUN go build -o sso ./cmd/sso/main.go
EXPOSE 44044
ENTRYPOINT [ "./sso" ]