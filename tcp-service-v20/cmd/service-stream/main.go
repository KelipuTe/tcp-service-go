package main

import (
	"fmt"
	"log"
	tcp_service_v20 "tcp-service-go/tcp-service-v20"
	"tcp-service-go/tcp-service-v20/internal/protocol"
	"tcp-service-go/tcp-service-v20/internal/service"
	"tcp-service-go/tcp-service-v20/internal/tool/signal"
)

func main() {
	log.Println("version: ", tcp_service_v20.Version)

	p1service := service.NewTCPService(protocol.StreamStr, "127.0.0.1", 9501)
	p1service.SetName(fmt.Sprintf("%s-service", protocol.StreamStr))
	p1service.SetDebugStatusOn()
	p1service.Start()

	signal.WaitForShutdown()
}
