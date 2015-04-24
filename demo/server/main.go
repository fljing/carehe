// server project main.go
package main

import (
	"carehe/network"
	//"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

var (
	careheStatus bool
	server       *network.CareheServer
)

type DemoServer struct {
}

func (p DemoServer) Process(client *network.CareheClient, data []byte) {
	fmt.Println(string(data))
}

func (p DemoServer) OnError(client *network.CareheClient, err error) {
	fmt.Println(err.Error())
}

func (p DemoServer) Connect(client *network.CareheClient) {
	fmt.Println("New Connect")
}

func (p DemoServer) Disconnect(client *network.CareheClient) {
	fmt.Println("Disconnect")
}

func startCarehe() {
	careheStatus = true
	fmt.Println("启动Carehe服务器")
	demo := DemoServer{}
	server = network.NewCareheServer("tcp", ":8888", demo)
	server.Start()
}

func stopCarehe() {
	fmt.Println("正在停止Carehe服务器")
	server.Stop()
	fmt.Println("已停止Carehe服务器")
}

func start(w http.ResponseWriter, req *http.Request) {
	if careheStatus == false || !server.IsRun() {
		w.Write([]byte("Carehe服务器正在启动"))
		startCarehe()
	} else {
		w.Write([]byte("Carehe服务器正在运行"))
		fmt.Println("Carehe服务器正在运行")
	}
}

func stop(w http.ResponseWriter, req *http.Request) {
	if careheStatus == true && server.IsRun() {
		w.Write([]byte("Carehe服务器正在关闭"))
		stopCarehe()
	} else {
		w.Write([]byte("Carehe服务器未启动"))
		fmt.Println("Carehe服务器未启动")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/start", start)
	http.HandleFunc("/stop", stop)
	http.ListenAndServe(":8080", nil)
}
