package util

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"

	dlmrdb "github.com/gomodule/redigo/redis"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix/v3"
	"github.com/olivere/elastic"

	"github.com/ikeikeikeike/gocore/util/dlm"
	"github.com/ikeikeikeike/gocore/util/dsn"
	"github.com/ikeikeikeike/gocore/util/logger"
)

// DBConn returns current database established connection
func DBConn(env Environment) (*sql.DB, error) {
	return SelectDBConn(env.EnvString("DSN"))
}

// SelectDBConn can choose db connection
func SelectDBConn(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("it was unable to connect the DB. %s", err)
	}

	// db configuration
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// make sure connection available
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("it was unable to connect the DB: %s", err)
	}

	var ver string
	logger.D("%s", db.QueryRow("SELECT @@version").Scan(&ver))

	msg := "[INFO] the mysql connection established <%s>, version %s"
	logger.Printf(msg, strings.Join(strings.Split(dsn, "@")[1:], ""), ver)

	return db, nil
}

// ESConn returns established connection
func ESConn(env Environment) (*elastic.Client, error) {
	var op []elastic.ClientOptionFunc
	op = append(op, elastic.SetHttpClient(&http.Client{Timeout: 5 * time.Second}))
	op = append(op, elastic.SetURL(env.EnvString("ESURL")))
	op = append(op, elastic.SetSniff(false))
	op = append(op, elastic.SetErrorLog(log.New(os.Stderr, "[ELASTIC] ", log.LstdFlags)))

	if env.IsDebug() {
		op = append(op, elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]] ", log.LstdFlags)))
		op = append(op, elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)))
	}

	es, err := elastic.NewClient(op...)
	if err != nil {
		return nil, fmt.Errorf("uninitialized es client <%s>: %s", env.EnvString("ESURL"), err)
	}
	ver, err := es.ElasticsearchVersion(env.EnvString("ESURL"))
	if err != nil {
		return nil, fmt.Errorf("error got es version <%s>: %s", env.EnvString("ESURL"), err)
	}

	msg := "[INFO] the elasticsearch connection established <%s>, version %s"
	logger.Printf(msg, env.EnvString("ESURL"), ver)
	return es, nil
}

// RDBConn returns established connection
// This is duplicated. use RDBV3Conn instead.
func RDBConn(env Environment) (*pool.Pool, error) {
	df := func(args ...interface{}) pool.DialFunc {
		return func(network, addr string) (*redis.Client, error) {
			client, err := redis.DialTimeout(network, addr, 5*time.Second)
			if err != nil {
				return nil, err
			}
			// select db
			if err = client.Cmd("SELECT", args...).Err; err != nil {
				client.Close()
				return nil, err
			}

			return client, nil
		}
	}

	dr, err := dsn.Redis(env.EnvString("RDBURI"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis dsn <%s>: %s", env.EnvString("RDBURI"), err)
	}
	p, err := pool.NewCustom("tcp", dr.HostPort, 10, df(dr.DB))
	if err != nil {
		return nil, fmt.Errorf("uninitialized redis client <%s>: %s", env.EnvString("RDBURI"), err)
	}

	msg := "[INFO] the redis connection established <%s>, version UNKNOWN"
	logger.Printf(msg, env.EnvString("RDBURI"))

	return p, err
}

// RDBV3Conn returns established connection
func RDBV3Conn(env Environment) (*radix.Pool, error) {
	uri := env.EnvString("RDBURI")

	dr, err := dsn.Redis(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis dsn <%s>: %s", uri, err)
	}

	selectDB, err := strconv.Atoi(dr.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis db number <%s>: %s", uri, err)
	}

	// this is a ConnFunc which will set up a connection which is authenticated
	// and has a 1 minute timeout on all operations
	connFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(time.Second*10),
			radix.DialSelectDB(selectDB),
		)
	}

	p, err := radix.NewPool("tcp", dr.HostPort, 10, radix.PoolConnFunc(connFunc))
	if err != nil {
		return nil, fmt.Errorf("uninitialized redis client <%s>: %s", uri, err)
	}

	msg := "[INFO] the redis@v3 connection established <%s>, version UNKNOWN"
	logger.Printf(msg, uri)

	return p, err
}

// DLMConn returns distributed lock manager pool
func DLMConn(env Environment) (*dlm.DLM, error) {
	dr, err := dsn.Redis(env.EnvString("DLMURI"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DLM dsn <%s>: %s", env.EnvString("DLMURI"), err)
	}

	pool := &dlmrdb.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (dlmrdb.Conn, error) {
			c, err := dlmrdb.Dial("tcp", dr.HostPort)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", dr.DB); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c dlmrdb.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	defer conn.Close()

	if _, err := dlmrdb.String(conn.Do("PING")); err != nil {
		return nil, fmt.Errorf("uninitialized DLM client <%s>: %s", env.EnvString("DLMURI"), err)
	}

	msg := "[INFO] the DLM(distributed lock) connection established <%s>, version UNKNOWN"
	logger.Printf(msg, env.EnvString("DLMURI"))

	return &dlm.DLM{Pool: pool}, nil
}

// BQConn returns err
func BQConn(env Environment) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, env.EnvString("GCProject"))
	if err != nil {
		return fmt.Errorf("there is no project in bigquery <%s>: %s", env.EnvString("GCProject"), err)
	}
	defer client.Close()

	msg := "[INFO] the bigquery connection established <%s>"
	logger.Printf(msg, env.EnvString("GCProject"))
	return nil
}
