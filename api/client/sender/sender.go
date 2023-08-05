package sender

import (
	"context"
	"github.com/MeysamBavi/go-broker/api/client/collector"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

const (
	responseReceiveChannelBuffer = 2000
)

type Sender struct {
	Host          string
	PublishStream <-chan *pb.PublishRequest
	FetchStream   <-chan *pb.FetchRequest
	Connections   int
	Verbose       bool
}

func (s *Sender) Start() <-chan collector.ResponseLog {
	if s.PublishStream == nil {
		ch := make(chan *pb.PublishRequest)
		close(ch)
		s.PublishStream = ch
	}
	if s.FetchStream == nil {
		ch := make(chan *pb.FetchRequest)
		close(ch)
		s.FetchStream = ch
	}

	responseLogStream := make(chan collector.ResponseLog, responseReceiveChannelBuffer)
	var wg sync.WaitGroup
	for i := 0; i < s.Connections; i++ {
		wg.Add(1)
		go func() {
			s.handleConnection(responseLogStream)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(responseLogStream)
	}()

	return responseLogStream
}

func (s *Sender) handleConnection(responseLogStream chan<- collector.ResponseLog) {
	conn, err := grpc.Dial(s.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatal("could not connect to server", err)
	}

	client := pb.NewBrokerClient(conn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		var publishWg sync.WaitGroup
		for request := range s.PublishStream {
			publishWg.Add(1)
			go s.sendPublishRequest(&publishWg, client, ctx, request, responseLogStream)
		}
		cancel()
		publishWg.Wait()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		var fetchWg sync.WaitGroup
		for request := range s.FetchStream {
			fetchWg.Add(1)
			go s.sendFetchRequest(&fetchWg, client, ctx, request, responseLogStream)
		}
		cancel()
		fetchWg.Wait()
		wg.Done()
	}()

	wg.Wait()
}

func (s *Sender) sendPublishRequest(wg *sync.WaitGroup, client pb.BrokerClient, ctx context.Context, request *pb.PublishRequest, receiveTimeStream chan<- collector.ResponseLog) {
	res, err := client.Publish(ctx, request)
	if s.Verbose {
		if err != nil {
			log.Printf("could not publish: %v\n", err)
		} else {
			log.Printf("published %q to subject %q, got id = %d\n", request.GetBody(), request.GetSubject(), res.GetId())
		}
	}
	receiveTimeStream <- collector.ResponseLog{
		At:    time.Now(),
		Error: err,
	}
	wg.Done()
}

func (s *Sender) sendFetchRequest(wg *sync.WaitGroup, client pb.BrokerClient, ctx context.Context, request *pb.FetchRequest, receiveTimeStream chan<- collector.ResponseLog) {
	res, err := client.Fetch(ctx, request)
	if s.Verbose {
		if err != nil {
			log.Printf("could not fetch: %v\n", err)
		} else {
			log.Printf("fetched id = %d from subject %q, got %q\n", request.GetId(), request.GetSubject(), res.GetBody())
		}
	}
	receiveTimeStream <- collector.ResponseLog{
		At:    time.Now(),
		Error: err,
	}
	wg.Done()
}
