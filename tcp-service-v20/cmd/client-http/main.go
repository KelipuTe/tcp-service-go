package main

import (
  "demo_golang/tcp_service_v1"
  "demo_golang/tcp_service_v1/internal/client"
  "demo_golang/tcp_service_v1/internal/protocol"
  "fmt"
  "log"
)

func main() {
  log.Println("version: ", tcp_service_v1.Version)

  p1client := client.NewTCPClient(protocol.HTTPStr, "127.0.0.1", 9501)
  p1client.SetName(fmt.Sprintf("%s-client", protocol.HTTPStr))
  p1client.SetDebugStatusOn()
  p1client.Start()
}
