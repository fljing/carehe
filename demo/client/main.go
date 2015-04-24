// client project main.go
package main

import (
	"carehe/network"
	"fmt"
	"runtime"
	"time"
)

type DemoClient struct {
}

func (p DemoClient) Process(client *network.CareheClient, data []byte) {
	fmt.Println(string(data))
}

func (p DemoClient) OnError(client *network.CareheClient, err error) {
	fmt.Println(err.Error())
}

func (p DemoClient) Connect(client *network.CareheClient) {
	fmt.Println("New Connect")
}

func (p DemoClient) Disconnect(client *network.CareheClient) {
	fmt.Println("Disconnect")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Hello World!")
	demo := DemoClient{}
	client := network.NewCareheClient(demo)
	client.Connect("tcp", "127.0.0.1:8888")
	client.Write([]byte("ffffffff"))
	client.Write([]byte("kkkkkkkkkk"))

	for i := 0; i < 100; i++ {
		dd := fmt.Sprintf("jjjjj%d", i)
		client.Write([]byte(dd))
		//time.Sleep(1 * time.Second)
	}
	time.Sleep(1 * time.Second)
	client.Close()
	time.Sleep(10 * time.Second)
}
