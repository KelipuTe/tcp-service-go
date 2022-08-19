package client

import (
  "demo_golang/tcp_service_v1/internal/protocol"
  "demo_golang/tcp_service_v1/internal/protocol/http"
  "demo_golang/tcp_service_v1/internal/protocol/stream"
  "demo_golang/tcp_service_v1/internal/protocol/websocket"
  "errors"
  "fmt"
  "io"
  "net"
)

const (
  // 接收缓冲区最大大小
  RecvBufferMax uint64 = 10 * 1048576
)

// TCPConnection TCP 连接
type TCPConnection struct {
  // 连接状态，详见 RunStatus 开头的常量
  runStatus uint8

  // TCP 连接所属 TCP 客户端
  p1client *TCPClient

  // 协议名称
  protocolName string
  // protocol.Protocol
  p1protocol protocol.Protocol

  // net.Conn
  p1conn net.Conn
  // 接收缓冲区
  sli1recvBuffer []byte
  // 接收缓冲区最大大小
  recvBufferMax uint64
  // 接收缓冲区当前大小
  recvBufferNow uint64
}

// NewTCPConnection 创建 TCPConnection
func NewTCPConnection(p1client *TCPClient, p1netConn net.Conn) *TCPConnection {
  p1tcpConn := &TCPConnection{
    runStatus:      RunStatusOn,
    p1client:       p1client,
    protocolName:   "",
    p1protocol:     nil,
    p1conn:         p1netConn,
    sli1recvBuffer: make([]byte, RecvBufferMax),
    recvBufferMax:  RecvBufferMax,
    recvBufferNow:  0,
  }

  p1tcpConn.protocolName = p1client.protocolName

  switch p1tcpConn.protocolName {
  case protocol.HTTPStr:
    p1tcpConn.p1protocol = http.NewHTTP()
  case protocol.StreamStr:
    p1tcpConn.p1protocol = stream.NewStream()
  case protocol.WebSocketStr:
    p1tcpConn.p1protocol = websocket.NewWebSocket()
  }

  return p1tcpConn
}

// IsRun TCP 连接是不是正在运行
func (p1this *TCPConnection) IsRun() bool {
  return RunStatusOn == p1this.runStatus
}

// TCPClient.IsDebug
func (p1this *TCPConnection) IsDebug() bool {
  return p1this.p1client.IsDebug()
}

// HandleConnection 处理连接
func (p1this *TCPConnection) HandleConnection(deferFunc func()) {
  defer func() {
    deferFunc()
  }()

  // 连上服务端之后，发送测试消息
  switch p1this.protocolName {
  case protocol.HTTPStr:
    // 发送 HTTP 消息
    resp := http.NewResponse()
    resp.SetStatusCode(http.StatusOk)
    respStr := resp.MakeResponse(fmt.Sprintf("this is %s.", p1this.p1client.name))
    p1this.SendMsg([]byte(respStr))
  case protocol.StreamStr:
    // 发送自定义字节流消息
    t1p1protocol := p1this.p1protocol.(*stream.Stream)
    t1p1protocol.SetDecodeMsg(fmt.Sprintf("this is %s.", p1this.p1client.name))
    p1this.SendMsg([]byte{})
  case protocol.WebSocketStr:
    // 发送 WebSocket 握手消息
    t1p1protocol := p1this.p1protocol.(*websocket.WebSocket)
    sli1respMsg, _ := t1p1protocol.MakeHandShakeReq()
    p1this.WriteData(sli1respMsg)
  }

  // 发送完了之后等待服务端响应
  for p1this.IsRun() {
    byteNum, err := p1this.p1conn.Read(p1this.sli1recvBuffer[p1this.recvBufferNow:])

    if p1this.IsDebug() {
      fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleConnection.byteNum: %d", p1this.p1client.name, byteNum))
    }

    if nil != err {
      if err == io.EOF {
        // 对端关闭了连接
        p1this.CloseConnection()
        return
      }
      p1this.p1client.OnClientError(p1this.p1client, err)
      return
    }

    p1this.recvBufferNow += uint64(byteNum)

    if p1this.IsDebug() {
      fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleConnection.recvBufferNow: %d", p1this.p1client.name, p1this.recvBufferNow))
      fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleConnection.sli1recvBuffer:", p1this.p1client.name))
      fmt.Println(string(p1this.sli1recvBuffer[0:p1this.recvBufferNow]))
    }

    p1this.HandleBuffer()
  }
}

// HandleBuffer 处理缓冲区
func (p1this *TCPConnection) HandleBuffer() {
  sli1Copy := p1this.sli1recvBuffer[0:p1this.recvBufferNow]
  for p1this.recvBufferNow > 0 {
    firstMsgLength, err := p1this.p1protocol.FirstMsgLength(sli1Copy)
    sli1firstMsg := p1this.sli1recvBuffer[0:firstMsgLength]

    switch p1this.protocolName {
    case protocol.HTTPStr:
      // HTTP 1.1 协议的消息，解析之后直接输出
      t1p1protocol := p1this.p1protocol.(*http.HTTP)
      t1p1protocol.Decode(sli1firstMsg)

      if p1this.IsDebug() {
        fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleBuffer.StreamStr.Decode: ", p1this.p1client.name))
        fmt.Println(fmt.Sprintf("%+v", t1p1protocol))
      }
      p1this.p1client.OnConnRequest(p1this)
    case protocol.StreamStr:
      // 自定义 Stream 协议的消息，解析之后直接输出
      t1p1protocol := p1this.p1protocol.(*stream.Stream)
      t1p1protocol.Decode(sli1firstMsg)

      if p1this.IsDebug() {
        fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleBuffer.StreamStr.Decode: ", p1this.p1client.name))
        fmt.Println(fmt.Sprintf("%+v", t1p1protocol))
      }
      p1this.p1client.OnConnRequest(p1this)
    case protocol.WebSocketStr:
      // WebSocket 协议的消息，需要判断是握手消息还是测试消息
      t1p1protocol := p1this.p1protocol.(*websocket.WebSocket)
      t1p1protocol.Decode(sli1firstMsg)

      if t1p1protocol.IsHandshakeStatusNo() {
        // 握手消息，校验一下服务端响应的握手消息
        err = t1p1protocol.CheckHandShakeResp()
        if err == nil {
          t1p1protocol.SetHandshakeStatusYes()
          t1p1protocol.SetDecodeMsg(fmt.Sprintf("this is %s.", p1this.p1client.name))
          p1this.SendMsg([]byte{})
        } else {
          p1this.CloseConnection()
        }
      } else {
        // 测试消息，解析之后直接输出
        if p1this.IsDebug() {
          fmt.Println(fmt.Sprintf("%s.TCPConnection.HandleBuffer.WebSocketStr.Decode: ", p1this.p1client.name))
          fmt.Println(fmt.Sprintf("%+v", t1p1protocol))
          p1this.p1client.OnConnRequest(p1this)
        }
      }
    }

    p1this.sli1recvBuffer = p1this.sli1recvBuffer[firstMsgLength:]
    // recvBufferNow 是 uint64 类型的，做减法的时候小心溢出
    if p1this.recvBufferNow <= firstMsgLength {
      p1this.recvBufferNow = 0
      break
    } else {
      p1this.recvBufferNow -= firstMsgLength
    }
  }
}

// SendMsg 发送数据
func (p1this *TCPConnection) SendMsg(sli1msg []byte) {
  switch p1this.protocolName {
  case protocol.TCPStr, protocol.HTTPStr:
    if p1this.IsDebug() {
      fmt.Println(fmt.Sprintf("%s.TCPConnection.SendMsg: ", p1this.p1client.name))
      fmt.Println(string(sli1msg))
    }
    p1this.WriteData(sli1msg)
  case protocol.StreamStr, protocol.WebSocketStr:
    t1sli1msg, _ := p1this.p1protocol.Encode()
    if p1this.IsDebug() {
      fmt.Println(fmt.Sprintf("%s.TCPConnection.SendMsg: ", p1this.p1client.name))
      fmt.Println(string(t1sli1msg))
    }
    p1this.WriteData(t1sli1msg)
  }
}

// WriteData 发送数据
func (p1this *TCPConnection) WriteData(sli1data []byte) (err error) {
  byteNum, err := p1this.p1conn.Write(sli1data)

  if p1this.IsDebug() {
    fmt.Println(fmt.Sprintf("%s.TCPConnection.WriteData.byteNum: %d", p1this.p1client.name, byteNum))
  }

  if nil != err {
    p1this.p1client.OnClientError(p1this.p1client, err)
    p1this.CloseConnection()
  }

  if byteNum != len(sli1data) {
    return errors.New("write byte != data length")
  }
  return nil
}

// CloseConnection 关闭连接
func (p1this *TCPConnection) CloseConnection() {
  p1this.runStatus = RunStatusOff
  p1this.recvBufferNow = 0
  p1this.p1client.OnConnClose(p1this)
  p1this.p1conn.Close()
}
