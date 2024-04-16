<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/neepooha/sso/raw/main/assets/images/logo-white.png">
    <img alt="logo" src="https://github.com/neepooha/sso/raw/main/assets/images/logo-black.png" width="40%">
  </picture>
</div>

<br><br>

<div align="center">
  
[![License](https://img.shields.io/badge/License-MIT-red)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/neepooha/sso)](https://goreportcard.com/report/github.com/neepooha/sso)
[![Go Reference](https://pkg.go.dev/badge/github.com/neepooha/url_shortener.svg)](https://pkg.go.dev/github.com/neepooha/url_shortener)
[![Build Status](https://github.com/neepooha/url_shortener/actions/workflows/deploy.yml/badge.svg)](https://github.com/neepooha/url_shortener/actions/workflows/deploy.yml)
<br>
</div>

## Features

This project is developed as a grps microservice for authentication, authorization and permissions.

### The project used:

* Standard CRUD operations of a database table
* JWT-based authentication
* Environment dependent application configuration management
* Structured logging with contextual information
* Error handling with proper error response generation
* Database migration
* Data validation
* Containerizing an application in Docker
 
### The kit uses the following Go packages:

* RPC framework: [grpc](google.golang.org/grpc)
* Database access: [pgx](https://github.com/jackc/pgx)
* Database migration: [golang-migrate](https://github.com/golang-migrate/migrate)
* Data validation: [go-playground validator](https://github.com/go-playground/validator)
* Logging: [log/slog](https://pkg.go.dev/golang.org/x/exp/slog)
* JWT: [jwt-go](https://github.com/dgrijalva/jwt-go)
* Config reader: [cleanv](github.com/ilyakaznacheev/cleanenv)  
* Env reader: [godotenv](github.com/joho/godotenv)

<div align="center">
This project is a microservice that works in conjunction with a SSO microservice. For full functionality, you need to have two microservices running.
  
[![SSO](https://github-readme-stats.vercel.app/api/pin/?username=neepooha&repo=url_shortener&border_color=7F3FBF&bg_color=0D1117&title_color=C9D1D9&text_color=8B949E&icon_color=7F3FBF)](https://github.com/neepooha/url_shortener)
[![SSO](https://github-readme-stats.vercel.app/api/pin/?username=neepooha&repo=protos&border_color=7F3FBF&bg_color=0D1117&title_color=C9D1D9&text_color=8B949E&icon_color=7F3FBF)](https://github.com/neepooha/protos)
<br>
</div>

## Getting Started

If this is your first time encountering Go, please follow [the instructions](https://golang.org/doc/install) to install Go on your computer. 
The project requires Go 1.21 or above.

[Docker](https://www.docker.com/get-started) is also needed if you want to try the kit without setting up your own database server.
The project requires Docker 17.05 or higher for the multi-stage build support.

Also for simple run commands i use [Taskfile](https://taskfile.dev/installation/). 

After installing Go, Docker and TaskFile, run the following commands to start experiencing:
```shell
## RUN SSO
# download the project
git clone [https://github.com/neepooha/url_shortener.git](https://github.com/neepooha/sso)
cd sso

# create config.env with that text:
> ./config.env {
CONFIG_PATH=./config/local.yaml
POSTGRES_DB=url
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypass
}

# start a PostgreSQL database server in a Docker container
task db-start

# run the SSO server
go run ./cmd/sso
```
Also, you can start project in dev mode. For that you need rename in config.env
"CONFIG_PATH=./config/local.yaml" to "CONFIG_PATH=./config/dev.yaml" in both projects
and run following commads:
```shell
# run the SSO server
cd sso/
docker compose up --build
```
SSO-grpc Server running at http://localhost:44044. The server provides the following endpoints:
#### auth
* `Regiter`: register new user in db
* `Login`: log in to the application
* `GETUserID`: get user ID by name

#### permissions
* `SetAdmin`: set exists user to admin in your app. You need be creator of app
* `DelAdmin`: delete exists user from admin in your app. You need be creator of app
* `IsAdmin`: is the user an admin by userID
* `IsCreator`: is the user a creator by userID

#### apps
* `SetApp`: set new app in db. You will be creator of the app
* `DelApp`: delete exists apps. You need be creator of app
* `UpdApp`: update app name and secret
* `GetAppID`: get app id by app name

## Project Layout
Project has the following project layout:
```
sso/
├── cmd/                       start of applications of the project
├── config/                    configuration files for different environments
├── deployment/                configuration for create daemon in linux
├── internal/                  private application and library code
│   ├── app/                   application assembly
│   ├── config/                configuration library
│   ├── domain/                models of apps and users
│   ├── grpc/                  grpc handlers
│   │   ├── apps/              handlers of apps
│   │   ├── auth/              handlers of auth
│   │   └── permissions/       handlers of permissions
│   ├── lib/                   additional functions for logging, error handling, migration
│   ├── services/              logics of handlers
│   │   ├── apps/              handlers of apps
│   │   ├── auth/              handlers of auth
│   │   └── permissions/       handlers of permissions
│   └── storage/               storage library
├── migrations/                migrations
└── config.env                 config for sercret variables
```
The top level directories `cmd`, `internal`, `lib` are commonly found in other popular Go projects, as explained in
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).

Within each feature package, code are organized in layers (grpc server, service, db), following the dependency guidelines
as described in the [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

### Updating Database Schema
for simple migration you can use the following commands
```shell
# For up migrations
task up

# For drop migrations
task drop

# Revert the last database migration.
tasl rollback
```
### Managing Configurations

The application configuration is represented in `internal/config/config.go`. When the application starts,
it loads the configuration from a configuration environment as well as environment variables. The path to the configuration environment
should be in the project root folder.

The `config` directory contains the configuration files named after different environments. For example,
`config/local.yml` corresponds to the local development environment and is used when running the application 
via `go run ./cmd/url-shortener`

You can keep secrets in local/dev confing, but do not keep secrets in the prud and in the configuration environment.
For set secret variable user github secrets and deploy.yaml.
