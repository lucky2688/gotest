package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"time"

	pb "github.com/lucky2688/gotest/pubsub-proto-ca/protobuf" //调用自己pubsub-proto-ca下的
)

type Authentication struct {
	User     string
	Password string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
	map[string]string, error,
) {
	return map[string]string{"user": a.User, "password": a.Password}, nil
}
func (a *Authentication) RequireTransportSecurity() bool {
	return false
}

// gRPC 客户端 发布
func main() {
	//在客户端就可以基于 CA 证书对服务器进行证书验证

	// 客户端证书（可选）
	certificate, err := tls.LoadX509KeyPair("./pubsub-proto-ca/client.crt", "./pubsub-proto-ca/client.key")
	if err != nil {
		log.Fatal("Load client.crt error:", err)
	}

	// 加载 CA 证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("./pubsub-proto-ca/ca.crt")
	if err != nil {
		log.Fatal("Load ca.crt error:", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	// TLS 配置
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   "server.grpc.io", // 必须与服务端证书中的 SAN 或 CN 匹配
		RootCAs:      certPool,
	})

	//加入token认证
	auth := Authentication{
		User:     "admin",
		Password: "123456",
	}

	// 建立连接
	conn, err := grpc.Dial("localhost:1234",
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&auth),
	)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	//发布程序
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

	// 调用SomeMethod程序
	helloClient := pb.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //上下文超时机制
	defer cancel()

	resp, err := helloClient.SomeMethod(ctx, &pb.HelloRequest{Name: "Lucky"})
	if err != nil {
		log.Fatalf("SomeMethod failed: %v", err)
	}
	fmt.Printf("SomeMethod Response: %s\n", resp.GetMessage())
}
