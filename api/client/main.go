package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"
)

var (
	host    = flag.String("host", "localhost:50043", "the host to connect to")
	subject = flag.String("subject", "alpha", "the subject of messages")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatal("could not connect to server", err)
	}

	client := pb.NewBrokerClient(conn)

	ctx := context.Background()
	messages, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		Subject: *subject,
	})
	if err != nil {
		log.Fatal("could not call Subscribe: ", err)
	}

	publishAndFetch(client, "neo", 11)
	publishAndFetch(client, "smith", 7)

	resp, streamErr := messages.Recv()
	for streamErr == nil {
		log.Printf("\t\tgot message %q\n", resp.Body)
		resp, streamErr = messages.Recv()
	}
	if streamErr == io.EOF {
		log.Println("got all messages")
	} else {
		log.Fatal("unknown error while reading stream: ", streamErr)
	}
}

func publishAndFetch(client pb.BrokerClient, prefix string, count int) {
	ids := make(chan int32, 1)
	go publishMessages(client, prefix, count, ids)
	go fetchMessages(client, ids)
}

func publishMessages(client pb.BrokerClient, prefix string, count int, ids chan<- int32) {
	defer close(ids)
	i := 0
	for ; count > 0; count-- {
		ctx := context.Background()
		body := fmt.Appendf(nil, "%s-%d", prefix, i)
		resp, err := client.Publish(ctx, &pb.PublishRequest{
			Subject:           *subject,
			Body:              body,
			ExpirationSeconds: 1,
		})
		if err != nil {
			log.Println("could not publish: ", err)
			return
		}
		log.Printf("published %q with id %d\n", body, resp.Id)
		ids <- resp.Id
		i++
		time.Sleep(500 * time.Millisecond)
	}
}

func fetchMessages(client pb.BrokerClient, ids <-chan int32) {
	ctx := context.Background()
	for id := range ids {
		response, err := client.Fetch(ctx, &pb.FetchRequest{
			Subject: *subject,
			Id:      id,
		})
		if err != nil {
			log.Printf("could not fetch id=%d: %v\n", id, err)
			return
		}
		log.Printf("fetched %q with id %d\n", response.Body, id)
	}
}
