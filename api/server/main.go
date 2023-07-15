package server

import (
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/internal/broker"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Execute(config Config) {
	//TODO: configure and use logger

	lis, err := net.Listen("tcp", config.Host)
	if err != nil {
		log.Fatal("could not listen", err)
	}

	s := grpc.NewServer()
	// TODO: configure broker settings
	pb.RegisterBrokerServer(s, NewServer(broker.NewModule()))

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
