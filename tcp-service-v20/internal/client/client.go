package client

import (
  "fmt"
  "net"
  "strconv"
  "sync"

  pkgErrors "github.com/pkg/errors"
)

const (
  RunStatusOff uint8 = iota // 服务（连接）关闭
  RunStatusOn               // 服务（连接）运行
)

const (
  DebugStatusOff uint8 = iota // debug 关
  DebugStatusOn               // debug 开
)

// TCPClient 默认属性值

const defaultName string = "default-client"

// TCPClient 默认方法

func defaultOnClientStart(p1service *TCPClient) {
  if p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnServiceStart", p1service.name))
  }
}

func defaultOnClientError(p1service *TCPClient, err error) {
  if p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnServiceError", p1service.name))
  }
  fmt.Println(fmt.Sprintf("%s", err))
}

func defaultOnConnConnect(p1conn *TCPConnection) {
  if p1conn.p1client.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnConnect", p1conn.p1client.name))
  }
}

func defaultOnConnRequest(p1conn *TCPConnection) {
  if p1conn.p1client.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnRequest", p1conn.p1client.name))
  }
}

func defaultOnConnClose(p1conn *TCPConnection) {
  if p1conn.p1client.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnClose", p1conn.p1client.name))
  }
}

// TCPClient TCP 客户端
type TCPClient struct {
  // name 客户端名称
  name string
  // runStatus 运行状态，详见 RunStatus 开头的常量
  runStatus uint8
  // debugStatus debug 开关状态，详见 DebugStatus 开头的常量
  debugStatus uint8

  // protocolName 协议名称
  protocolName string

  // address IP 地址
  address string
  // port 端口号
  port uint16

  // TCP 连接
  p1conn *TCPConnection

  // OnClientStart 客户端启动事件回调
  OnClientStart func(*TCPClient)
  // OnClientError 客户端错误事件回调
  OnClientError func(*TCPClient, error)
  // OnConnConnect TCP 连接，连接事件回调
  OnConnConnect func(*TCPConnection)
  // OnConnRequest TCP 连接，请求事件回调
  OnConnRequest func(*TCPConnection)
  // OnConnClose TCP 连接，关闭事件回调
  OnConnClose func(*TCPConnection)
}

// NewTCPClient 创建默认的 TCPClient
func NewTCPClient(protocolName string, address string, port uint16) *TCPClient {
  return &TCPClient{
    name:         defaultName,
    runStatus:    RunStatusOn,
    debugStatus:  DebugStatusOff,
    protocolName: protocolName,
    address:      address,
    port:         port,

    OnClientStart: defaultOnClientStart,
    OnClientError: defaultOnClientError,
    OnConnConnect: defaultOnConnConnect,
    OnConnRequest: defaultOnConnRequest,
    OnConnClose:   defaultOnConnClose,
  }
}

// SetName 设置客户端名称
func (p1this *TCPClient) SetName(name string) {
  p1this.name = name
}

// SetDebugStatusOn 打开 debug
func (p1this *TCPClient) SetDebugStatusOn() {
  p1this.debugStatus = DebugStatusOn
}

// IsDebug 是否是 debug 模式
func (p1this *TCPClient) IsDebug() bool {
  return DebugStatusOn == p1this.debugStatus
}

func (p1this *TCPClient) Start() {
  p1this.OnClientStart(p1this)

  p1conn, err := net.Dial("tcp4", p1this.address+":"+strconv.Itoa(int(p1this.port)))
  if nil != err {
    p1this.OnClientError(p1this, pkgErrors.WithMessage(err, "TCPClient.StartListen"))
    return
  }

  p1this.p1conn = NewTCPConnection(p1this, p1conn)
  p1this.OnConnConnect(p1this.p1conn)

  var wg sync.WaitGroup
  wg.Add(1)
  go p1this.p1conn.HandleConnection(wg.Done)
  wg.Wait()
}
