package main

import (
	"context"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"strings"
	"time"

	"github.com/moby/pubsub"
	"google.golang.org/grpc"

	pb "github.com/lucky2688/gotest/pubsub-proto/protobuf"
)

type PubsubService struct {
	//嵌入 UnimplementedPubsubServiceServer 结构体，这是从 protoc-gen-go-grpc v1.4.0+ 开始要求的一个标准写法
	pb.UnimplementedPubsubServiceServer
	pub *pubsub.Publisher
}

func NewPubsubService() *PubsubService {
	return &PubsubService{
		pub: pubsub.NewPublisher(100*time.Millisecond, 10),
	}
}

func (p *PubsubService) Publish(ctx context.Context, msg *pb.String) (*pb.String, error) {
	p.pub.Publish(msg.GetValue())
	return &pb.String{}, nil
}

func (p *PubsubService) Subscribe(filter *pb.String, stream pb.PubsubService_SubscribeServer) error {
	ch := p.pub.SubscribeTopic(func(v interface{}) bool {
		if val, ok := v.(string); ok {
			return strings.HasPrefix(val, filter.Value)
		}
		return false
	})

	for val := range ch {
		if err := stream.Send(&pb.String{Value: val.(string)}); err != nil {
			return err
		}
	}

	return nil
}

// gRPC 服务端
func main() {
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//grpcServer := grpc.NewServer()  无证书

	//有证书
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatal("TLS error:", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterPubsubServiceServer(grpcServer, NewPubsubService())

	log.Println("gRPC Pubsub server started on :1234")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/*
windows下用cmd或者powershell执行生成证书
自 Go 1.15+ 和 gRPC 使用的 crypto/x509 模块开始，证书校验默认要求包含 SAN
# 生成私钥
openssl genrsa -out server.key 2048
# 生成带 SAN 的证书
openssl req -x509 -new -nodes -key server.key -days 3650 -out server.crt -config cert.conf


到main.go和证书同一级目录下执行
PS C:\projectgo\gotest> cd .\pubsub-proto\
PS C:\projectgo\gotest\pubsub-proto> go run main.go
2025/07/24 19:08:56 gRPC Pubsub server started on :1234



以上这种方式，需要提前将服务器的证书告知客户端，这样客户端在连接服务器时才能进行对服务器证书认证。
在复杂的网络环境中，服务器证书的传输本身也是一个非常危险的问题。
如果在中间某个环节，服务器证书被监听或替换那么对服务器的认证也将不再可靠。

*/
