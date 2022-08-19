package user

import (
  "demo_golang/tcp_service_v2/internal/api"
  "demo_golang/tcp_service_v2/internal/protocol/stream"
  "encoding/json"
)

func (p1this *UserService) GetUserName(p1apipkg *api.APIPackage) {
  p1req := &api.ReqInUserName{}
  json.Unmarshal([]byte(p1apipkg.Data), p1req)

  if p1req.Id == 1 {
    p1req.Name = "aaa"
  } else if p1req.Id == 2 {
    p1req.Name = "bbb"
  }

  p1reqJson, _ := json.Marshal(p1req)
  p1apipkg.Type = api.TypeResponse
  p1apipkg.Data = string(p1reqJson)
  p1apipkgJson, _ := json.Marshal(p1apipkg)

  t1p1protocol := p1this.p1innerClient.GetTCPConn().GetProtocol().(*stream.Stream)
  t1p1protocol.SetDecodeMsg(string(p1apipkgJson))
  p1this.p1innerClient.GetTCPConn().SendMsg([]byte{})
}

func (p1this *UserService) GetUserLevel(p1apipkg *api.APIPackage) {
  p1req := &api.ReqInUserLevel{}
  json.Unmarshal([]byte(p1apipkg.Data), p1req)

  if p1req.Id == 1 {
    p1req.Level = 11
  } else if p1req.Id == 2 {
    p1req.Level = 22
  }

  p1reqJson, _ := json.Marshal(p1req)
  p1apipkg.Type = api.TypeResponse
  p1apipkg.Data = string(p1reqJson)
  p1apipkgJson, _ := json.Marshal(p1apipkg)

  t1p1protocol := p1this.p1innerClient.GetTCPConn().GetProtocol().(*stream.Stream)
  t1p1protocol.SetDecodeMsg(string(p1apipkgJson))
  p1this.p1innerClient.GetTCPConn().SendMsg([]byte{})
}
