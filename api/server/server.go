package server

import (
	"context"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var (
	errInternal    = status.Errorf(codes.Internal, "internal error")
	errUnavailable = status.Error(codes.Unavailable, broker.ErrUnavailable.Error())
)

type server struct {
	pb.UnimplementedBrokerServer
	broker broker.Broker
}

func NewServer(bk broker.Broker) pb.BrokerServer {
	return &server{
		broker: bk,
	}
}

func (s *server) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	body := string(request.GetBody())
	id, err := s.broker.Publish(ctx, request.GetSubject(), broker.Message{
		Body:       body,
		Expiration: time.Duration(request.GetExpirationSeconds()) * time.Second,
	})

	if err == nil {
		return &pb.PublishResponse{
			Id: int32(id),
		}, nil
	}

	if err == broker.ErrUnavailable {
		return nil, errUnavailable
	}

	//TODO: log error
	return nil, errInternal
}

func (s *server) Subscribe(request *pb.SubscribeRequest, subscribeServer pb.Broker_SubscribeServer) error {
	sub, err := s.broker.Subscribe(subscribeServer.Context(), request.GetSubject())

	if err == nil {
		for message := range sub {
			err := subscribeServer.Send(&pb.MessageResponse{
				Body: []byte(message.Body),
			})
			if err != nil {
				//TODO: log error
				return status.Errorf(codes.Internal, "could not send message: %v", err)
			}
		}
		return nil
	}

	if err == broker.ErrUnavailable {
		return errUnavailable
	}

	//TODO: log error
	return errInternal
}

func (s *server) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	id := int(request.GetId())
	message, err := s.broker.Fetch(ctx, request.GetSubject(), id)
	if err == nil {
		return &pb.MessageResponse{
			Body: []byte(message.Body),
		}, nil
	}

	if err == broker.ErrUnavailable {
		return nil, errUnavailable
	}

	if err == broker.ErrExpiredID || err == broker.ErrInvalidID {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument for id=%d; message expired or not found", id)
	}

	//TODO: log error
	return nil, errInternal
}
