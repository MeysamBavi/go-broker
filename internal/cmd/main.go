package cmd

import (
	"encoding/json"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/broker"
	"github.com/MeysamBavi/go-broker/internal/config"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/pkg/metrics"
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

	var msgStore store.Message
	switch {
	case cfg.Store.UseInMemory:
		msgStore = store.NewInMemoryMessage(store.GetDefaultTimeProvider())
	case cfg.Store.UseCassandra:
		msgStore, err = store.NewCassandra(cfg.Store.Cassandra)
		if err != nil {
			log.Fatal("could not connect to cassandra: ", err)
		}
	}

	if cfg.Metrics.Enabled {
		go metrics.RunServer(cfg.Metrics)
	}

	s := grpc.NewServer()
	module := broker.NewModuleWithStore(msgStore)
	pb.RegisterBrokerServer(s, server.NewServer(module))

	log.Printf("server listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
