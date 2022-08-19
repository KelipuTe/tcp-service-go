package service

import (
  goErrors "errors"
  "fmt"
  "log"
  "net"
  "os"
  "runtime"
  "strconv"

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

// TCPService 默认属性值

const defaultName string = "default-service"

// TCPService 默认方法

func defaultOnServiceStart(p1service *TCPService) {
  if p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnServiceStart", p1service.name))
  }
}

func defaultOnServiceError(p1service *TCPService, err error) {
  if p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnServiceError", p1service.name))
  }
  fmt.Println(fmt.Sprintf("%s", err))
}

func defaultOnConnConnect(p1conn *TCPConnection) {
  if p1conn.p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnConnect", p1conn.p1service.name))
  }
}

func defaultOnConnRequest(p1conn *TCPConnection) {
  if p1conn.p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnRequest", p1conn.p1service.name))
  }
}

func defaultOnConnClose(p1conn *TCPConnection) {
  if p1conn.p1service.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.OnConnClose", p1conn.p1service.name))
  }
}

// TCPService TCP 服务端
type TCPService struct {
  // name 服务端名称
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
  // p1listener net.Listener
  p1listener net.Listener

  // mapConnPool TCP 连接（TCPConnection）池
  mapConnPool map[string]*TCPConnection
  // maxConnNum TCP 连接，最大连接数
  maxConnNum uint32
  // nowConnNum TCP 连接，当前连接数
  nowConnNum uint32

  // OnServiceStart 服务端启动事件回调
  OnServiceStart func(*TCPService)
  // OnServiceError 服务端错误事件回调
  OnServiceError func(*TCPService, error)
  // OnConnConnect TCP 连接，连接事件回调
  OnConnConnect func(*TCPConnection)
  // OnConnRequest TCP 连接，请求事件回调
  OnConnRequest func(*TCPConnection)
  // OnConnClose TCP 连接，关闭事件回调
  OnConnClose func(*TCPConnection)
}

// NewTCPService 创建默认的 TCPService
func NewTCPService(protocolName string, address string, port uint16) *TCPService {
  return &TCPService{
    name:         defaultName,
    runStatus:    RunStatusOn,
    debugStatus:  DebugStatusOff,
    protocolName: protocolName,
    address:      address,
    port:         port,

    mapConnPool: make(map[string]*TCPConnection),
    maxConnNum:  1024,
    nowConnNum:  0,

    OnServiceStart: defaultOnServiceStart,
    OnServiceError: defaultOnServiceError,
    OnConnConnect:  defaultOnConnConnect,
    OnConnRequest:  defaultOnConnRequest,
    OnConnClose:    defaultOnConnClose,
  }
}

// SetName 设置服务名称
func (p1this *TCPService) SetName(name string) {
  p1this.name = name
}

// IsRun 服务是不是正在运行
func (p1this *TCPService) IsRun() bool {
  return RunStatusOn == p1this.runStatus
}

// SetDebugStatusOn 打开 debug
func (p1this *TCPService) SetDebugStatusOn() {
  p1this.debugStatus = DebugStatusOn
}

// IsDebug 是否是 debug 模式
func (p1this *TCPService) IsDebug() bool {
  return DebugStatusOn == p1this.debugStatus
}

// Start 服务启动
func (p1this *TCPService) Start() {
  p1this.StartInfo()

  t1address := p1this.address + ":" + strconv.Itoa(int(p1this.port))
  listener, err := net.Listen("tcp4", t1address)
  if nil != err {
    p1this.OnServiceError(p1this, pkgErrors.WithMessage(err, fmt.Sprintf("%s.Start", p1this.name)))
    return
  }

  p1this.p1listener = listener
  defer p1this.p1listener.Close()

  p1this.OnServiceStart(p1this)
  p1this.StartListen()
}

// StartInfo 输出服务配置和环境参数
func (p1this *TCPService) StartInfo() {
  log.Println("runtime.GOOS=", runtime.GOOS)
  log.Println("runtime.NumCPU()=", runtime.NumCPU())
  log.Println("runtime.Version()=", runtime.Version())
  log.Println("os.Getpid()=", os.Getpid())
}

// StartListen 开始监听
func (p1this *TCPService) StartListen() {
  for p1this.IsRun() {
    // net.Listener.Accept，系统调用，获取 TCP 连接
    p1netConn, err := p1this.p1listener.Accept()
    if nil != err {
      p1this.OnServiceError(p1this, pkgErrors.WithMessage(err, fmt.Sprintf("%s.StartListen", p1this.name)))
      return
    }
    // 判断 TCP 连接当前数量是否超过最大连接数
    if p1this.nowConnNum >= p1this.maxConnNum {
      err = goErrors.New("nowConnNum >= maxConnNum")
      p1this.OnServiceError(p1this, pkgErrors.WithMessage(err, fmt.Sprintf("%s.StartListen", p1this.name)))
    }

    p1TCPConn := NewTCPConnection(p1this, p1netConn)
    p1this.AddConnection(p1TCPConn)
    p1this.OnConnConnect(p1TCPConn)
    go p1TCPConn.HandleConnection()
  }
}

// AddConnection 添加连接
func (p1this *TCPService) AddConnection(p1tcpConn *TCPConnection) {
  // 用 Linux C 编码时，可以通过 socket 的文件描述符区分 TCP 连接
  // 在 go 中也可以获得文件描述符，但是文件描述符不是唯一的，不能用于区分
  if p1this.IsDebug() {
    fd, err := p1tcpConn.p1conn.(*net.TCPConn).File()
    fmt.Println("net.TCPConn.File", fd.Fd(), err)
  }

  addrStr := p1tcpConn.p1conn.RemoteAddr().String()
  p1this.nowConnNum++
  p1this.mapConnPool[addrStr] = p1tcpConn

  if p1this.IsDebug() {
    fmt.Println("net.TCPConn.RemoteAddr.String", addrStr)
  }
}

// DeleteConnection 移除连接
func (p1this *TCPService) DeleteConnection(p1tcpConn *TCPConnection) {
  addrStr := p1tcpConn.p1conn.RemoteAddr().String()
  if _, ok := p1this.mapConnPool[addrStr]; ok {
    delete(p1this.mapConnPool, addrStr)
    p1this.nowConnNum--
  }
}
