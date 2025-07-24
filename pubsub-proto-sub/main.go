package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

// gRPC 客户端 订阅
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
