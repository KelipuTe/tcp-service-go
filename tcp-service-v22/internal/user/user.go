package user

import (
  "demo_golang/tcp_service_v2/internal/api"
  "demo_golang/tcp_service_v2/internal/client"
  "demo_golang/tcp_service_v2/internal/protocol"
  "demo_golang/tcp_service_v2/internal/protocol/stream"
  "encoding/json"
)

var P1UserService *UserService

func init() {
  P1UserService = &UserService{}
}

// HandlerFunc 路由对应的处理方法
type HandlerFunc func(*api.APIPackage)

// UserService 用户服务
type UserService struct {
  // 需要一个内部 TCP 客户端连接到 gateway
  p1innerClient *client.TCPClient
  // 路由表，记录路由和处理方法
  mapRoute map[string]HandlerFunc
  // 路由表，发送给 gateway 的
  sli1Route []string
}

// SetInnerClient 设置内部 TCP 客户端
func (p1this *UserService) SetInnerClient(p1client *client.TCPClient) {
  p1this.p1innerClient = p1client
}

// RegisteServiceProvider 向 gateway 发送服务提供者的注册信息
func (p1this *UserService) RegisteServiceProvider() {
  // 定义路由表
  p1this.mapRoute = map[string]HandlerFunc{
    api.APIUserName:  p1this.GetUserName,
    api.APIUserLevel: p1this.GetUserLevel,
  }
  for key := range p1this.mapRoute {
    p1this.sli1Route = append(p1this.sli1Route, key)
  }

  // 拼装数据
  p1apipkg := &api.APIPackage{}
  p1apipkg.Type = api.TypeRequest
  p1apipkg.Action = api.ActionRegisteServiceProvider
  t1data := &api.ReqInRegisteServiceProvider{
    Name:      p1this.p1innerClient.GetName(),
    Sli1Route: p1this.sli1Route,
  }
  t1dataJson, _ := json.Marshal(t1data)
  p1apipkg.Data = string(t1dataJson)
  p1apipkgJson, _ := json.Marshal(p1apipkg)

  // 发送数据
  protocolName := p1this.p1innerClient.GetTCPConn().GetProtocolName()
  switch protocolName {
  case protocol.StreamStr:
    t1p1protocol := p1this.p1innerClient.GetTCPConn().GetProtocol().(*stream.Stream)
    t1p1protocol.SetDecodeMsg(string(p1apipkgJson))
    p1this.p1innerClient.GetTCPConn().SendMsg([]byte{})
  }
}

// DispatchRequest 处理 gateway 发送过来的请求
func (p1this *UserService) DispatchRequest(p1conn *client.TCPConnection) {
  p1apipkg := &api.APIPackage{}
  msg := p1conn.GetProtocol().(*stream.Stream).GetDecodeMsg()
  json.Unmarshal([]byte(msg), p1apipkg)

  switch p1apipkg.Type {
  case api.TypeRequest:
    switch p1apipkg.Action {
    case api.ActionPing:
      p1apipkg.Type = api.TypeRequest
      p1apipkg.Action = api.ActionPong
      p1apipkg.Data = p1conn.GetNetConnRemoteAddr()
      p1apipkgJson, _ := json.Marshal(p1apipkg)

      t1p1protocol := p1this.p1innerClient.GetTCPConn().GetProtocol().(*stream.Stream)
      t1p1protocol.SetDecodeMsg(string(p1apipkgJson))
      p1this.p1innerClient.GetTCPConn().SendMsg([]byte{})
    default:
      // 从路由表中查找处理函数，APIPackage.Action 就是 api
      t1func := p1this.mapRoute[p1apipkg.Action]
      t1func(p1apipkg)
    }
  case api.TypeResponse:
  }
}
