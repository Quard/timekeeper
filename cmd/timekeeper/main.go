package main

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jessevdk/go-flags"

	"github.com/Quard/timekeeper/internal/storage"
	"github.com/Quard/timekeeper/internal/timekeeper_rpc"
)

var opts struct {
	Bind      string `short:"b" long:"bind" env:"BIND" default:"localhost:5020"`
	MongoURI  string `long:"mongo-uri" env:"MONGO_URI" default:"mongodb://localhost:27017/timekeeper"`
	SentryDSN string `long:"sentry-dsn" env:"SENTRY_DSN"`
}

func main() {
	parser := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash)
	if _, err := parser.Parse(); err != nil {
		log.Fatal(err)
	}

	sentry.Init(sentry.ClientOptions{Dsn: opts.SentryDSN})
	defer sentry.Flush(5 * time.Second)

	stor := storage.NewMongoStorage(opts.MongoURI)
	srv := timekeeper_rpc.NewTimeKeeperRPC(timekeeper_rpc.Opts{Bind: opts.Bind}, stor)
	srv.Run()
}
