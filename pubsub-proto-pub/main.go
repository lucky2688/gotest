package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

// gRPC 客户端 发布
func main() {
	//conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())  无证书

	//基于服务器的证书和服务器名字就可以对服务器进行验证
	creds, err := credentials.NewClientTLSFromFile(
		"../pubsub-proto/server.crt", "server.grpc.io",
	)
	if err != nil {
		log.Fatal("TLS error:", err)
	}

	conn, err := grpc.Dial("localhost:1234",
		grpc.WithTransportCredentials(creds),
	)

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
