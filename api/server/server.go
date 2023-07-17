package server

import (
	"context"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"github.com/MeysamBavi/go-broker/pkg/metrics"
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
	broker         broker.Broker
	metricsHandler metrics.Handler
	timeProvider   store.TimeProvider
}

func NewServer(bk broker.Broker, metricsHandler metrics.Handler, timeProvider store.TimeProvider) pb.BrokerServer {
	return &server{
		broker:         bk,
		metricsHandler: metricsHandler,
		timeProvider:   timeProvider,
	}
}

func (s *server) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	success := false
	callTime := s.timeProvider.GetCurrentTime()
	defer func() {
		latency := s.timeProvider.GetCurrentTime().Sub(callTime)
		s.metricsHandler.ReportPublishLatency(latency)
		s.metricsHandler.IncPublishCallCount(success)
	}()

	body := string(request.GetBody())
	id, err := s.broker.Publish(ctx, request.GetSubject(), broker.Message{
		Body:       body,
		Expiration: time.Duration(request.GetExpirationSeconds()) * time.Second,
	})

	if err == nil {
		success = true
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
	success := false
	defer func() {
		s.metricsHandler.IncPublishCallCount(success)
	}()

	sub, err := s.broker.Subscribe(subscribeServer.Context(), request.GetSubject())

	if err == nil {
		s.metricsHandler.IncActiveSubscribers()
		defer s.metricsHandler.DecActiveSubscribers()

		for message := range sub {
			err := subscribeServer.Send(&pb.MessageResponse{
				Body: []byte(message.Body),
			})
			if err != nil {
				//TODO: log error
				return status.Errorf(codes.Internal, "could not send message: %v", err)
			}
		}

		success = true
		return nil
	}

	if err == broker.ErrUnavailable {
		return errUnavailable
	}

	//TODO: log error
	return errInternal
}

func (s *server) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	success := false
	callTime := s.timeProvider.GetCurrentTime()
	defer func() {
		latency := s.timeProvider.GetCurrentTime().Sub(callTime)
		s.metricsHandler.ReportFetchLatency(latency)
		s.metricsHandler.IncFetchCallCount(success)
	}()

	id := int(request.GetId())
	message, err := s.broker.Fetch(ctx, request.GetSubject(), id)
	if err == nil {
		success = true
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
