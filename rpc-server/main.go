package main

import (
	"io"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// 接口名称常量
const HelloServiceName = "HelloService"

// 定义接口
type HelloServiceInterface interface {
	Hello(request string, reply *string) error
}

// 实现结构体
type HelloService struct{}

// 实现接口方法
// 方法必须满足 Go 语言的 RPC 规则：方法只能有两个可序列化的参数，其中第二个参数是指针类型，并且返回一个 error 类型，同时必须是公开的方法
func (h *HelloService) Hello(request string, reply *string) error {
	*reply = "Hello, " + request
	return nil
}

// 注册服务
func RegisterHelloService(svc HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, svc)
}

func main() {
	// 注册服务
	err := RegisterHelloService(new(HelloService))
	if err != nil {
		log.Fatal("Register error:", err)
	}

	/*
		// 开始监听
		listener, err := net.Listen("tcp", ":1234")
		if err != nil {
			log.Fatal("ListenTCP error:", err)
		}
		defer listener.Close()

		fmt.Println("RPC server listening on :1234")

		// 支持多个连接
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}
			//go rpc.ServeConn(conn)
			//基于 json 的jsonrpc服务
			go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
		}*/

	//基于http的jsonrpc服务
	http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		var conn io.ReadWriteCloser = struct {
			io.Writer
			io.ReadCloser
		}{
			ReadCloser: r.Body,
			Writer:     w,
		}
		defer r.Body.Close()

		rpc.ServeRequest(jsonrpc.NewServerCodec(conn))
	})

	http.ListenAndServe(":1234", nil)
}

//curl localhost:1234/jsonrpc -X POST --data '{"method":"HelloService.Hello","params":["kitty"],"id":0}'
//{"id":0,"result":"Hello, kitty","error":null}
