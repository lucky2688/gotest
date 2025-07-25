package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/moby/pubsub"
	"google.golang.org/grpc"

	pb "github.com/lucky2688/gotest/pubsub-proto-ca/protobuf" //注意自己的路径
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

// 加入token认证
type Authentication struct {
	User     string
	Password string
}
type helloService struct {
	pb.UnimplementedHelloServiceServer
	auth *Authentication
}

func (a *Authentication) Auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing credentials")
	}

	var appid, appkey string

	if val, ok := md["user"]; ok {
		appid = val[0]
	}
	if val, ok := md["password"]; ok {
		appkey = val[0]
	}

	if appid != a.User || appkey != a.Password {
		return grpc.Errorf(codes.Unauthenticated, "invalid token")
	}

	return nil
}

// gRPC 服务端
func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd error: %v", err)
	}
	fmt.Println("Working dir:", dir)
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	certificate, err := tls.LoadX509KeyPair("pubsub-proto-ca/server.crt", "pubsub-proto-ca/server.key")
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("pubsub-proto-ca/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert, // NOTE: this is optional!
		ClientCAs:    certPool,
	})
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterPubsubServiceServer(grpcServer, NewPubsubService())

	// 注册 HelloService
	auth := &Authentication{
		User:     "admin",
		Password: "123456",
	}
	pb.RegisterHelloServiceServer(grpcServer, &helloService{auth: auth}) // &helloService{} 不开启token认证

	log.Println("gRPC Pubsub server started on :1234")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *helloService) SomeMethod(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received from client: %s", in.GetName())

	// 示例：可选权限认证
	if s.auth != nil {
		if err := s.auth.Auth(ctx); err != nil {
			return nil, err
		}
	}

	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

/*
生成代码  cmd,git bash，powershell都可以执行，重开
$ cd ../pubsub-proto-ca
lucky@DESKTOP-OG5FQBI MINGW64 /c/projectgo/gotest/pubsub-proto-ca (main)
$ protoc --go_out=. --go-grpc_out=. hello.proto


客户端就可以基于 CA 证书对服务器进行证书验证

为了避免证书的传递过程中被篡改，可以通过一个安全可靠的根证书分别对服务器和客户端的证书进行签名。
这样客户端或服务器在收到对方的证书后可以通过根证书进行验证证书的有效性。

windows下用cmd或者powershell执行生成证书
自 Go 1.15+ 和 gRPC 使用的 crypto/x509 模块开始，证书校验默认要求包含 SAN

根证书的生成方式和自签名证书的生成方式类似：
1. 生成 CA 证书（用于签发其他证书）
openssl genrsa -out ca.key 2048
openssl req -x509 -new -key ca.key -days 3650 -out ca.crt -config ca.conf

然后是重新对服务器端证书进行签名：
2. 生成 Server 证书（由 CA 签发，带 SAN）
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -extensions v3_req -extfile server.conf


3. 生成 Client 证书（可选，用于双向认证）
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -config client.conf
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 3650 -extensions v3_req -extfile client.conf

签名的过程中引入了一个新的以. csr 为后缀名的文件，它表示证书签名请求文件。
在证书签名完成之后可以删除. csr 文件。

所有程序统一在执行目录C:\projectgo\gotest下执行
PS C:\projectgo\gotest> go run .\pubsub-proto-ca\main.go
API server listening at: 127.0.0.1:55474
Working dir: C:\projectgo\gotest
2025/07/25 17:28:00 gRPC Pubsub server started on :1234

*/
