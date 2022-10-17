package main

import (
	"fmt"
	"log"
	tcp_service_v20 "tcp-service-go/tcp-service-v20"
	"tcp-service-go/tcp-service-v20/internal/client"
	"tcp-service-go/tcp-service-v20/internal/protocol"
)

func main() {
	log.Println("version: ", tcp_service_v20.Version)

	p1client := client.NewTCPClient(protocol.WebSocketStr, "127.0.0.1", 9501)
	p1client.SetName(fmt.Sprintf("%s-client", protocol.WebSocketStr))
	p1client.SetDebugStatusOn()
	p1client.Start()
}
