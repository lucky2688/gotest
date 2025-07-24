package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

// gRPC 客户端 订阅
func main() {
	//在客户端就可以基于 CA 证书对服务器进行证书验证
	certificate, err := tls.LoadX509KeyPair("../pubsub-proto-ca/client.crt", "../pubsub-proto-ca/client.key")
	if err != nil {
		log.Fatal("Load client.crt error:", err)
	}
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../pubsub-proto-ca/ca.crt")
	if err != nil {
		log.Fatal("Load ca.crt error:", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   "server.grpc.io", // NOTE: this is required!
		RootCAs:      certPool,
	})

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
