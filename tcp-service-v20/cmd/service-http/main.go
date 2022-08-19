package main

import (
  "demo_golang/tcp_service_v1"
  "demo_golang/tcp_service_v1/internal/protocol"
  "demo_golang/tcp_service_v1/internal/service"
  "demo_golang/tcp_service_v1/internal/tool/signal"
  "fmt"
  "log"
)

func main() {
  log.Println("version: ", tcp_service_v1.Version)

  p1service := service.NewTCPService(protocol.HTTPStr, "127.0.0.1", 9501)
  p1service.SetName(fmt.Sprintf("%s-service", protocol.HTTPStr))
  p1service.SetDebugStatusOn()
  p1service.Start()

  signal.WaitForShutdown()
}
