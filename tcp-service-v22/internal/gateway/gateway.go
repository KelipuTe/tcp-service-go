package gateway

import (
  "demo_golang/tcp_service_v2/internal/api"
  "demo_golang/tcp_service_v2/internal/service"
  "encoding/json"
  "fmt"
  "strconv"
  "time"
)

const defaultName string = "default-gateway"

const (
  DebugStatusOff uint8 = iota // debug 关
  DebugStatusOn               // debug 开
)

var P1gateway *Gateway

func init() {
  P1gateway = &Gateway{
    name:              defaultName,
    debugStatus:       DebugStatusOff,
    mapInnerConnPool:  make(map[string][]*service.TCPConnection),
    mapInnerConnCount: make(map[string]uint64),
    mapConnToPing:     make(map[string]*service.TCPConnection),
    mapOpenConn:       make(map[string]*service.TCPConnection),
  }
}

// Gateway 服务
type Gateway struct {
  // name Gateway 服务名称
  name string
  // debugStatus debug 开关状态，详见 DebugStatus 开头的常量
  debugStatus uint8

  // p1innerService 需要一个内部 TCP 服务端为服务提供者提供服务。
  p1innerService *service.TCPService

  // mapInnerConnPool 不同服务提供者的 TCP 连接池。
  // 一个服务提供者注册之后，在这里会变成多个键值对。
  // 服务提供者提供的每个 api 都会对应服务提供者的 TCP 连接。
  mapInnerConnPool map[string][]*service.TCPConnection
  // mapInnerConnCount 记录每个 api 被调用的次数，用于实现简单的负载均衡。
  mapInnerConnCount map[string]uint64
  // mapConnToPing 需要保持心跳的 TCP 连接
  mapConnToPing map[string]*service.TCPConnection

  // mapOpenConn 外部请求的 TCP 连接。
  // 一个外部请求连接上之后，在这里保存 IP 和 TCP 连接的关系，用于发送响应数据。
  mapOpenConn map[string]*service.TCPConnection
}

// SetDebugStatusOn 打开 debug
func (p1this *Gateway) SetDebugStatusOn() {
  p1this.debugStatus = DebugStatusOn
}

// IsDebug 是否是 debug 模式
func (p1this *Gateway) IsDebug() bool {
  return DebugStatusOn == p1this.debugStatus
}

// SetInnerService 设置内部 TCP 服务端
func (p1this *Gateway) SetInnerService(p1service *service.TCPService) {
  p1this.p1innerService = p1service
}

func (p1this *Gateway) StartPingConn() {
  for {
    for t1addr, t1conn := range p1this.mapConnToPing {
      p1apipkg := &api.APIPackage{}
      p1apipkg.Id = t1addr
      p1apipkg.Type = api.TypeRequest
      p1apipkg.Action = api.ActionPing
      p1apipkg.Data = strconv.FormatInt(time.Now().Unix(), 10)

      p1this.SendInnerResponse(t1conn, p1apipkg)
    }
    time.Sleep(10 * time.Second)
  }
}

// RegisteServiceProvider 接收服务提供者的注册信息
func (p1this *Gateway) RegisteServiceProvider(p1conn *service.TCPConnection, p1apipkg *api.APIPackage) {
  p1req := &api.ReqInRegisteServiceProvider{}
  json.Unmarshal([]byte(p1apipkg.Data), p1req)

  // 服务提供者的每个 api，都要生成一个键值对
  // 这样查找服务提供者的逻辑，就可以直接用 api 来查找
  for _, api := range p1req.Sli1Route {
    if p1this.IsDebug() {
      fmt.Println(fmt.Sprintf("%s.RegisteServiceProvider, api: %s, ip: %s", p1this.name, api, p1conn.GetNetConnRemoteAddr()))
    }

    sli1bizService, ok := p1this.mapInnerConnPool[api]
    if ok {
      sli1bizService = append(sli1bizService, p1conn)
    } else {
      sli1bizService = []*service.TCPConnection{p1conn}
    }
    p1this.mapInnerConnPool[api] = sli1bizService
    p1this.mapInnerConnCount[api] = 0
  }

  // 添加服务提供者的连接到心跳列表
  p1this.mapConnToPing[p1conn.GetNetConnRemoteAddr()] = p1conn
}

// GetInnerConn 获取 api 对应的服务提供者的 TCP 连接
func (p1this *Gateway) GetInnerConn(api string) *service.TCPConnection {
  sli1conn, ok := p1this.mapInnerConnPool[api]
  if !ok {
    return nil
  }
  connNum := len(sli1conn)
  if connNum <= 0 {
    return nil
  }
  p1this.mapInnerConnCount[api] = p1this.mapInnerConnCount[api] + 1
  connIndex := p1this.mapInnerConnCount[api] % uint64(connNum)
  return p1this.mapInnerConnPool[api][connIndex]
}

// DeleteServiceProvider 移除服务提供者
func (p1this *Gateway) DeleteServiceProvider(p1conn *service.TCPConnection) {
  // 将服务提供者的连接移出心跳列表
  t1addr := p1conn.GetNetConnRemoteAddr()
  delete(p1this.mapConnToPing, t1addr)
  // 因为注册的时候服务提供者的每个 api 都会单独注册
  // 所以移除的时候也需要针对每个 api 去移除
  for api, sli1Conn := range p1this.mapInnerConnPool {
    for index, t1p1Conn := range sli1Conn {
      // 在池子里找到 ip 对应的那个 TCP 连接
      if t1addr == t1p1Conn.GetNetConnRemoteAddr() {
        sli1Conn := append(sli1Conn[0:index], sli1Conn[index+1:]...)
        p1this.mapInnerConnPool[api] = sli1Conn

        if p1this.IsDebug() {
          fmt.Println(fmt.Sprintf("%s.DeleteServiceProvider, api: %s, ip: %s", p1this.name, api, t1addr))
        }

        break
      }
    }
  }
}
