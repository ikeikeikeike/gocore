# Go Core

This is common package powered by [uber-go/dig](https://github.com/uber-go/dig) dependency injection toolkit.


>
> N/A go@1.11
>
> A   go@1.12.9
>

This system force required to: You **have to must** have `GO111MODULE` environment variable that activates `go mod` module.

Therefore, it set into.


```php
$ export GO111MODULE=on
```

or use direnv.

```php
$ brew install direnv
```

## Installation

Download Project

```php
$ go get -u github.com:ikeikeikeike/gocore
$ go get -u github.com:ikeikeikeike/gocore/v1
$ go get -u github.com:ikeikeikeike/gocore@rom5kdxv4kfq0uhfq1hfq4
```

## Usage

Make sure what tasks are existing.

```php
$ make help
```

Update Golang dependencies. This command would be help you when something went wrong happended.

```php
$ make gomodule
```

## Components


### Environment Variable


- [Environment](/util/env.go)


### Connection establishment

- [Connection](/util/conn.go)


### Utils

- [Struct Merge](/util/structs)
- [RPC](/util/rpc)
- [Repository](/util/repo)
- [Data Source Name](/util/dsn)
- [Graceful Restart](/util/graceful)
- [Sentry Logger](/util/logger)
- [Crypto,JWT](/util/crypto)
- [Distributed lock manager](/util/dlm)


### DATA I/O


- [File,S3,GCS](/data/storage)
- [Redis](/data/rdb)
- [Elasticsearch](/data/search)
- [BigQuery](/data/bq)


## Setup

Inherit util.Environment into your Environment struct.

```go
package main

import (
  "strings"

  "github.com/gigawattio/metaflector"
  "github.com/spf13/cast"

  "github.com/ikeikeikeike/gocore/util"
)

type (
  Environment interface {
    util.Environment
  }

  Env struct {
    // DSN is mysql data source name
    DSN string `envconfig:"GOCORE_API_DSN" default:"root:@tcp(127.0.0.1:3306)/tablename?parseTime=true"`

    // RDBURI is set server host and port with db number, that's like DSN
    RDBURI string `envconfig:"GOCORE_API_RDBURI" default:"redis://127.0.0.1:6379/10"`

    // DLMURI is set distributed lock server host and port with db number, that's like DSN
    DLMURI string `envconfig:"GOCORE_API_DLMURI" default:"redis://127.0.0.1:6379/9"`

    // URI is set server host and port, which means the same as FQDN
    ESURL string `envconfig:"GOCORE_API_ESURL" default:"http://127.0.0.1:9200"`

    // FURI is storage uri e.g. s3://data_bucket/path/data.flac or file://
    FURI string `envconfig:"GOCORE_API_FURI" default:"file://./storage/data.flac"`

    // BQGSURI is bigquery loader storage
    BQGSURI string `envconfig:"GOCORE_API_BQGSURI" default:"gs://gocore-bigquery-development/gocore/table.json"`

    // Debug controls
    Debug string `envconfig:"GOCORE_API_DEBUG" default:""` // debug|pprof|something

    // GCProject: GOOGLE_PROJECT, GCLOUD_PROJECT
    GCProject string `envconfig:"GOOGLE_PROJECT" default:""`

    // SentryDSN: SENTRY_DSN
    SentryDSN string `envconfig:"SENTRY_DSN" default:""`
  }
)

// IsSentry returns
func (e *Env) IsSentry() bool {
  return e.SentryDSN != ""
}

// IsDebug returns
func (e *Env) IsDebug() bool {
  return !e.IsProd() || e.Debug == "debug"
}

// EnvString returns as stirng
func (e *Env) EnvString(prop string) string {
  v, err := cast.ToStringE(metaflector.Get(e, prop))
  if err != nil {
    logger.Cretical("EnvString failed prop `%v`: %s", prop, err)
  }
  return v
}

// EnvInt returns as int
func (e *Env) EnvInt(prop string) int {
  v, err := cast.ToIntE(metaflector.Get(e, prop))
  if err != nil {
    logger.Cretical("EnvInt failed prop `%v`: %s", prop, err)
  }
  return v
}

```


Define Components as you like which is [uber-go/dig](https://github.com/uber-go/dig) way.


```go
package main

import (
  "go.uber.org/dig"

  "github.com/kelseyhightower/envconfig"
  "github.com/volatiletech/sqlboiler/boil"

  "github.com/ikeikeikeike/gocore/data/search"
  "github.com/ikeikeikeike/gocore/data/storage"
  "github.com/ikeikeikeike/gocore/util/logger"
  "github.com/ikeikeikeike/gocore/util"
)

func initInject(di *dig.Container) {
  var deps = []interface{}{
    initDB,
    initDLM,
    initES,
    initEnv,
    initRDB,
  }

  for _, dep := range deps {
    if err := di.Provide(dep); err != nil {
      logger.Panicf("failed to process root injection: %s", err)
    }
  }

  // Inject Core
  search.Inject(di)
  storage.Inject(di)
}

func initEnv() Environment {
  var env Environment = &Env{}
  if err := envconfig.Process("", env); err != nil {
    logger.Panicf("failed to get env var: %s", err)
  }

  logger.SetDebug(env.IsDebug())
  logger.SetSentry(env.IsSentry())
  // boil.DebugMode = !env.IsProd()

  return env
}

func initCoreEnv(env Environment) util.Environment {
  return env
}


func initDB(env util.Environment) *sql.DB {
  boil.DebugMode = !env.IsProd()

  // apidb
  db, err := util.DBConn(env)
  if err != nil {
    logger.Panicf("failed to get DBConn: %s", err)
  }

  return db
}

// elasticsearch
func initES(env util.Environment) *elastic.Client {
  es, err := util.ESConn(env) // defer es.Close()
  if err != nil {
    logger.Panicf("failed to get ESConn: %s", err)
  }
  return es
}


// redis
func initRDB(env util.Environment) *pool.Pool {
  db, err := util.RDBConn(env)
  if err != nil {
    logger.Panicf("failed to get RDBConn: %s", err)
  }

  return db
}

// distributed lock manager
func initDLM(env util.Environment) *dlm.DLM {
  // Redis
  db, err := util.DLMConn(env)
  if err != nil {
    logger.Panicf("failed to get DLMConn: %s", err)
  }

  return db
}

```

Invoke Components.

```go
package main

import (
  "database/sql"
  "log"
  "net"
  "os"
  "os/signal"
  "syscall"

  "go.uber.org/dig"
  "google.golang.org/grpc"

  _ "github.com/go-sql-driver/mysql"

  "github.com/olivere/elastic"
  "github.com/mediocregopher/radix.v2/pool"

  "github.com/ikeikeikeike/gocore/data/storage"
  "github.com/ikeikeikeike/gocore/util/dlm"
  "github.com/ikeikeikeike/gocore/util/graceful"
  "github.com/ikeikeikeike/gocore/util/logger"
)

func main() {
  di := dig.New()

  initInject(di)

  if err := di.Invoke(runServer); err != nil {
    log.Fatalf("could not run the application: %+v", err)
  }
}

// RunServerIn
//
// Declare all of interfaces which you will use.
//
type runServerIn struct {
  dig.In

  Env     Environment
  DB      *sql.DB
  DLM     *dlm.DLM
  ES      *elastic.Client
  RDB     *pool.Pool
  Storage storage.Storage
}

// runServer returns
func runServer(in runServerIn) {
  // a api server
  uri, err := url.Parse(in.Env.EnvString("URI"))
  if err != nil {
    logger.Panicf("failed to get parse api uri: %s", err)
  }

  errors := make(chan error)

  go func(rt *grpc.Server) {
    lis, err := net.Listen("tcp", uri.Host)
    if err != nil {
      logger.Panic("faild to listen: %v", err)
    }

    logger.Info("start grpc server: %s", uri.Host)

    errors <- rt.Serve(lis)
  }(in.Grpc)

  q := make(chan os.Signal)
  signal.Notify(q, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

  go func() {
    <-q

    // XXX: What is the best order to close connection?
    logger.Info("%s waiting remain hooks to closing...", uri.Host)
    graceful.Shutdown()

    logger.Info("%s waiting database to closing...", uri.Host)
    in.DB.Close()

    logger.Info("%s waiting distributed lock to closing...", uri.Host)
    in.DLM.Close()

    logger.Info("%s waiting redis to closing...", uri.Host)
    in.RDB.Empty()

    logger.Info("%s graceful stopping a grpc server...", uri.Host)
    in.Grpc.GracefulStop()
  }()

  if err := <-errors; err != nil {
    logger.Panicf("auba-api returned non-nil error on launch server: %v", err)
  }
}
```
