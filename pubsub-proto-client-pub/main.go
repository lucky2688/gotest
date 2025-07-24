package main

import (
	"context"
	"google.golang.org/grpc"
	"log"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

// gRPC 客户端 发布
func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	client := pb.NewPubsubServiceClient(conn)

	_, err = client.Publish(
		context.Background(), &pb.String{Value: "golang: hello Go"},
	)
	if err != nil {
		log.Fatal("Publish failed:", err)
	}
	_, err = client.Publish(
		context.Background(), &pb.String{Value: "docker: hello Docker"},
	)
	if err != nil {
		log.Fatal("Publish failed:", err)
	}

	log.Println("Messages published successfully")
}
