package main

import (
  "demo_golang/tcp_service_v2"
  "demo_golang/tcp_service_v2/internal/client"
  "demo_golang/tcp_service_v2/internal/protocol"
  "demo_golang/tcp_service_v2/internal/tool/signal"
  "demo_golang/tcp_service_v2/internal/user"
  "fmt"
  "log"
)

var p1innerClient *client.TCPClient

func main() {
  log.Println("version: ", tcp_service_v2.Version)

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
