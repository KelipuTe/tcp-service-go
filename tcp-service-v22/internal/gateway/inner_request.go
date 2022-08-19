package gateway

import (
  "demo_golang/tcp_service_v2/internal/api"
  "demo_golang/tcp_service_v2/internal/protocol/http"
  "demo_golang/tcp_service_v2/internal/protocol/stream"
  "demo_golang/tcp_service_v2/internal/service"
  "encoding/json"
  "fmt"
)

// DispatchInnerRequest 处理内部服务的请求
func (p1this *Gateway) DispatchInnerRequest(p1conn *service.TCPConnection) {
  p1apipkg := &api.APIPackage{}
  msg := p1conn.GetProtocol().(*stream.Stream).GetDecodeMsg()
  json.Unmarshal([]byte(msg), p1apipkg)

  switch p1apipkg.Type {
  case api.TypeRequest:
    switch p1apipkg.Action {
    case api.ActionRegisteServiceProvider:
      p1this.RegisteServiceProvider(p1conn, p1apipkg)

      p1apipkg.Type = api.TypeResponse
      p1apipkg.Data = "success."
      p1this.SendInnerResponse(p1conn, p1apipkg)
    }
  case api.TypeResponse:
    switch p1apipkg.Action {
    case api.ActionPong:
      if p1this.IsDebug() {
        fmt.Println(fmt.Sprintf("%s.TCPConnection.ActionPong: ip: %s", p1this.name, p1conn.GetNetConnRemoteAddr()))
      }
    default:
      resp := http.NewResponse()
      resp.SetStatusCode(http.StatusOk)
      respStr := resp.MakeResponse(p1apipkg.Data)

      t1p1connection := p1this.mapOpenConn[p1apipkg.Id]
      t1p1connection.SendMsg([]byte(respStr))
      t1p1connection.CloseConnection()
      // 移除失效的外部请求的 TCP 连接
      delete(p1this.mapOpenConn, p1apipkg.Id)
    }
  }
}

// SendInnerResponse 向内部服务发送响应
func (p1this *Gateway) SendInnerResponse(p1conn *service.TCPConnection, p1apipkg *api.APIPackage) {
  p1apipkgJson, _ := json.Marshal(p1apipkg)

  t1p1protocol := p1conn.GetProtocol().(*stream.Stream)
  t1p1protocol.SetDecodeMsg(string(p1apipkgJson))
  p1conn.SendMsg([]byte{})
}
