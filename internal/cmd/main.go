package cmd

import (
	"encoding/json"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/broker"
	"github.com/MeysamBavi/go-broker/internal/config"
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
		log.Fatal("could not listen", err)
	}

	s := grpc.NewServer()
	// TODO: configure broker settings
	// TODO: create different instances of store based on configs
	pb.RegisterBrokerServer(s, server.NewServer(broker.NewModule()))

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
