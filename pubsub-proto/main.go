package main

import (
	"context"
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
	grpcServer := grpc.NewServer()
	pb.RegisterPubsubServiceServer(grpcServer, NewPubsubService())

	log.Println("gRPC Pubsub server started on :1234")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
