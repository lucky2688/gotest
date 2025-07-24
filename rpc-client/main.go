package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// 接口名称常量
const HelloServiceName = "HelloService"

type HelloServiceClient struct {
	*rpc.Client
}

func DialHelloService(network, address string) (*HelloServiceClient, error) {
	/*client, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}*/

	//基于 json 编码
	conn, err := net.Dial(network, address)
	if err != nil {
		log.Fatal("net.Dial:", err)
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	return &HelloServiceClient{Client: client}, nil
}

func (p *HelloServiceClient) Hello(request string, reply *string) error {
	return p.Client.Call(HelloServiceName+".Hello", request, reply)
}

func main() {
	/*client, err := DialHelloService("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Hello("kitty", &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)*/

	//HTTP POST JSON请求
	req := map[string]interface{}{
		"method": "HelloService.Hello",
		"params": []interface{}{"kitty"},
		"id":     1,
	}

	body, _ := json.Marshal(req)

	resp, err := http.Post("http://localhost:1234/jsonrpc", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Server response:", string(respBody))

}
