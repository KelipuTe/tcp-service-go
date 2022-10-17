package main

import (
	"fmt"
	"log"
	tcp_service_v22 "tcp-service-go/tcp-service-v22"
	"tcp-service-go/tcp-service-v22/internal/client"
	"tcp-service-go/tcp-service-v22/internal/protocol"
	"tcp-service-go/tcp-service-v22/internal/tool/signal"
	"tcp-service-go/tcp-service-v22/internal/user"
)

var p1innerClient *client.TCPClient

func main() {
	log.Println("version: ", tcp_service_v22.Version)

	p1innerClient := client.NewTCPClient(protocol.StreamStr, "127.0.0.1", 9501)
	p1innerClient.SetName(fmt.Sprintf("%s-client-user", protocol.StreamStr))
	p1innerClient.SetDebugStatusOn()

	p1innerClient.OnClientStart = func(p1client *client.TCPClient) {
		if p1client.IsDebug() {
			fmt.Println(fmt.Sprintf("%s.OnServiceStart", p1client.GetName()))
		}
		user.P1UserService.SetInnerClient(p1client)
	}

	p1innerClient.OnConnConnect = func(p1conn *client.TCPConnection) {
		if p1innerClient.IsDebug() {
			fmt.Println(fmt.Sprintf("%s.OnConnConnect", p1innerClient.GetName()))
		}
		user.P1UserService.RegisteServiceProvider()
	}

	p1innerClient.OnConnRequest = func(p1conn *client.TCPConnection) {
		if p1innerClient.IsDebug() {
			fmt.Println(fmt.Sprintf("%s.OnConnConnect", p1innerClient.GetName()))
		}
		user.P1UserService.DispatchRequest(p1conn)
	}
	p1innerClient.Start()

	signal.WaitForShutdown()
}
