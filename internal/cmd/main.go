package cmd

import (
	"encoding/json"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/broker"
	"github.com/MeysamBavi/go-broker/internal/config"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/metrics"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Execute() {
	cfg := config.Load()
	{
		cfgJson, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("config: %s\n", cfgJson)
	}

	//TODO: configure and use logger
	lis, err := net.Listen("tcp", cfg.Server.Host)
	if err != nil {
		log.Fatal("could not listen: ", err)
	}

	tracerProvider, shutdown := tracing.NewTracerProvider(cfg.Tracing)
	defer shutdown()

	var sequenceStore store.Sequence
	sequenceStore = store.NewInMemorySequence()
	sequenceStore = store.SequenceWithTracing(sequenceStore, tracerProvider)

	var msgStore store.Message
	switch {
	case cfg.Store.UseInMemory:
		msgStore = store.NewInMemoryMessage(store.GetDefaultTimeProvider())
	case cfg.Store.UseCassandra:
		msgStore, err = store.NewCassandra(cfg.Store.Cassandra, tracerProvider)
		if err != nil {
			log.Fatal("could not connect to cassandra: ", err)
		}
	case cfg.Store.UsePostgres:
		msgStore, err = store.NewPostgres(cfg.Store.Postgres, sequenceStore, store.GetDefaultTimeProvider(), tracerProvider)
		if err != nil {
			log.Fatal("could not connect to postgres: ", err)
		}
	}
	msgStore = store.MessageWithTracing(msgStore, tracerProvider)

	subsStore := store.NewInMemorySubscriber()
	subsStore = store.SubscriberWithTracing(subsStore, tracerProvider)

	var metricsHandler metrics.Handler
	if cfg.Metrics.Enabled {
		metricsHandler = metrics.NewPrometheusHandler()
		go metrics.RunServer(cfg.Metrics)
	} else {
		metricsHandler = metrics.NewEmptyHandler()
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
	)
	module := broker.NewModuleWithStores(msgStore, subsStore)
	module = broker.WithTracing(module, tracerProvider)
	pb.RegisterBrokerServer(s, server.NewServer(module, metricsHandler, store.GetDefaultTimeProvider()))

	log.Printf("server listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
