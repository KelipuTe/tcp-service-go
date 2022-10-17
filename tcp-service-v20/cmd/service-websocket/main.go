package main

import (
	"fmt"
	"log"
	tcp_service_v20 "tcp-service-go/tcp-service-v20"
	"tcp-service-go/tcp-service-v20/internal/protocol"
	"tcp-service-go/tcp-service-v20/internal/service"
	"tcp-service-go/tcp-service-v20/internal/tool/signal"
)

// 可以用 EasySwoole-WebSocket在线测试工具 测试
// http://www.easyswoole.com/wstool.html
// 也可以直接用 JavaScript 的 WebSocket 工具
func main() {
	log.Println("version: ", tcp_service_v20.Version)

	p1service := service.NewTCPService(protocol.WebSocketStr, "127.0.0.1", 9501)
	p1service.SetName(fmt.Sprintf("%s-service", protocol.WebSocketStr))
	p1service.SetDebugStatusOn()
	p1service.Start()

	signal.WaitForShutdown()
}
