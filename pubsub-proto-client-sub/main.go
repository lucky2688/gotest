package main

import (
	"context"
	"google.golang.org/grpc"
	"log"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

// gRPC 客户端 订阅
func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	client := pb.NewPubsubServiceClient(conn)

	stream, err := client.Subscribe(context.Background(), &pb.String{Value: "golang"})
	if err != nil {
		log.Fatal(err)
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("Receive error:", err)
		}
		log.Println("Received:", msg.Value)
	}
}
