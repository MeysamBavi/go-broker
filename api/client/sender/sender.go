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
}

func (s *Sender) Start() <-chan collector.ResponseLog {
	publishStream := s.PublishStream
	if publishStream == nil {
		ch := make(chan *pb.PublishRequest)
		close(ch)
		publishStream = ch
	}
	fetchStream := s.FetchStream
	if fetchStream == nil {
		ch := make(chan *pb.FetchRequest)
		close(ch)
		fetchStream = ch
	}

	responseLogStream := make(chan collector.ResponseLog, responseReceiveChannelBuffer)
	var wg sync.WaitGroup
	for i := 0; i < s.Connections; i++ {
		wg.Add(1)
		go func() {
			handleConnection(s.Host, publishStream, fetchStream, responseLogStream)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(responseLogStream)
	}()

	return responseLogStream
}

func handleConnection(host string, publishStream <-chan *pb.PublishRequest, fetchStream <-chan *pb.FetchRequest, responseLogStream chan<- collector.ResponseLog) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		for request := range publishStream {
			publishWg.Add(1)
			go sendPublishRequest(&publishWg, client, ctx, request, responseLogStream)
		}
		cancel()
		publishWg.Wait()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		var fetchWg sync.WaitGroup
		for request := range fetchStream {
			fetchWg.Add(1)
			go sendFetchRequest(&fetchWg, client, ctx, request, responseLogStream)
		}
		cancel()
		fetchWg.Wait()
		wg.Done()
	}()

	wg.Wait()
}

func sendPublishRequest(wg *sync.WaitGroup, client pb.BrokerClient, ctx context.Context, request *pb.PublishRequest, receiveTimeStream chan<- collector.ResponseLog) {
	_, err := client.Publish(ctx, request)
	receiveTimeStream <- collector.ResponseLog{
		At:    time.Now(),
		Error: err,
	}
	wg.Done()
}

func sendFetchRequest(wg *sync.WaitGroup, client pb.BrokerClient, ctx context.Context, request *pb.FetchRequest, receiveTimeStream chan<- collector.ResponseLog) {
	_, err := client.Fetch(ctx, request)
	receiveTimeStream <- collector.ResponseLog{
		At:    time.Now(),
		Error: err,
	}
	wg.Done()
}
