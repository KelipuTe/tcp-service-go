package websocket

import (
  "crypto/md5"
  "crypto/sha1"
  "demo_golang/tcp_service_v1/internal/protocol"
  "demo_golang/tcp_service_v1/internal/protocol/http"
  "encoding/base64"
  "errors"
  "fmt"
  "strings"
)

const (
  handshakeStatusNo  uint8 = iota // 未握手
  handshakeStatusYes              // 已握手
)

const (
  encodeTypeUseMusk uint8 = iota // 用 Masking-key 解码
  encodeTypeNoMusk               // 不用 Masking-key 编码
)

const (
  opcodeText   uint8 = 0x01 // 文本帧
  opcodeBinary uint8 = 0x02 // 二进制帧
  opcodeClose  uint8 = 0x08 // 连接断开
  opcodePing   uint8 = 0x09 // ping
  opcodePong   uint8 = 0x0A // pong
)

var (
  // 报文不完整（没接收全）
  ErrDataIncomplete = errors.New("websocket sli1data incomplete.")
  // 连接已关闭
  ErrConnectionIsClosed = errors.New("websocket connection is closed.")
)

var _ protocol.Protocol = &WebSocket{}

// WebSocket 协议
// https://datatracker.ietf.org/doc/rfc6455/
// https://www.rfc-editor.org/rfc/rfc6455
type WebSocket struct {
  // 编码类型，详见 encodeType 开头的常量
  encodeType uint8

  // p1HttpInner 握手阶段要用 HTTP 协议
  p1HttpInner *http.HTTP

  // SecWebSocketKey sec-websocket-key
  SecWebSocketKey string

  // handshakeStatus 握手状态，详见 handshakeStatus 开头的常量
  handshakeStatus uint8

  // FIN，1 bit
  // 0（不是消息的最后一个分片）；1（这是消息的最后一个分片）；
  fin bool
  // opcode，4 bit
  opcode uint8
  // MASK，1 bit
  // 0（没有 Masking-key）；1（有 Masking-key）；
  mask bool
  // Payload len，7 bit
  payloadLen8 uint8
  // Extended payload length，16 bit，if payload len==126
  payloadLen16 uint16
  // Extended payload length，64 bit，if payload len==127
  payloadLen64 uint64
  // Masking-key，4 byte
  arr1MaskingKey [4]byte

  // headerLength 头部长度
  headerLength uint8
  // bodyLength 消息体长度
  bodyLength uint64

  // Sli1Msg 请求报文
  Sli1Msg []byte
  // DecodeMsg 解析后的数据
  DecodeMsg string
}

func NewWebSocket() *WebSocket {
  return &WebSocket{
    encodeType:      encodeTypeNoMusk,
    p1HttpInner:     http.NewHTTP(),
    handshakeStatus: handshakeStatusNo,
  }
}

func (p1this *WebSocket) FirstMsgLength(sli1recv []byte) (uint64, error) {
  if handshakeStatusNo == p1this.handshakeStatus {
    // 没有握手
    return p1this.p1HttpInner.FirstMsgLength(sli1recv)
  } else if handshakeStatusYes == p1this.handshakeStatus {
    // 已经握手
    recvLen := len(sli1recv)
    if recvLen < 2 {
      // 至少 2 个字节才能解析
      return 0, ErrDataIncomplete
    }

    // 取 FIN，第 1 个字节的第 1 位
    t1fin := sli1recv[0] & 0b10000000
    if t1fin == 0b10000000 {
      p1this.fin = true
    }

    // 取 opcode，第 1 个字节的后 4 位
    p1this.opcode = sli1recv[0] & 0b00001111
    if opcodeClose == p1this.opcode {
      return 0, ErrConnectionIsClosed
    }

    // 头部长度至少 2 字节
    p1this.headerLength = 2

    // 取 MASK，第 2 个字节的第 1 位
    mask := sli1recv[1] & 0b10000000
    if mask == 0b10000000 {
      p1this.mask = true
      // 有 Masking-key，头部长度 +4 字节
      p1this.headerLength += 4
    } else {
      p1this.mask = false
    }

    // 取 Payload len，第 2 个字节的后 7 位
    p1this.payloadLen8 = sli1recv[1] & 0b0111111
    if 126 == p1this.payloadLen8 {
      p1this.headerLength += 2
    } else if 127 == p1this.payloadLen8 {
      p1this.headerLength += 8
    }

    if int(p1this.headerLength) > recvLen {
      // 计算出来的报文长度大于接收缓冲区中数据长度
      return 0, ErrDataIncomplete
    }

    var msgLen uint64 = 0

    // 计算报文长度
    if 126 == p1this.payloadLen8 {
      // Payload len 为 126，需要扩展 2 个字节
      p1this.payloadLen16 = 0
      p1this.payloadLen16 |= uint16(sli1recv[2]) << 8
      p1this.payloadLen16 |= uint16(sli1recv[3]) << 0
      msgLen = uint64(p1this.headerLength) + uint64(p1this.payloadLen16)
    } else if 127 == p1this.payloadLen8 {
      // Payload len 为 127，需要扩展 8 个字节
      p1this.payloadLen64 |= uint64(sli1recv[2]) << 56
      p1this.payloadLen64 |= uint64(sli1recv[3]) << 48
      p1this.payloadLen64 |= uint64(sli1recv[4]) << 40
      p1this.payloadLen64 |= uint64(sli1recv[5]) << 32
      p1this.payloadLen64 |= uint64(sli1recv[6]) << 24
      p1this.payloadLen64 |= uint64(sli1recv[7]) << 16
      p1this.payloadLen64 |= uint64(sli1recv[8]) << 8
      p1this.payloadLen64 |= uint64(sli1recv[9]) << 0
      msgLen = uint64(p1this.headerLength) + uint64(p1this.payloadLen64)
    } else {
      msgLen = uint64(p1this.headerLength) + uint64(p1this.payloadLen8)
    }

    p1this.bodyLength = msgLen
    if msgLen > uint64(recvLen) {
      // 计算出来的报文长度大于接收缓冲区中数据长度
      return 0, ErrDataIncomplete
    }

    // 获取 Masking-Key
    if p1this.mask {
      p1this.arr1MaskingKey[0] = sli1recv[p1this.headerLength-4]
      p1this.arr1MaskingKey[1] = sli1recv[p1this.headerLength-3]
      p1this.arr1MaskingKey[2] = sli1recv[p1this.headerLength-2]
      p1this.arr1MaskingKey[3] = sli1recv[p1this.headerLength-1]
    }

    return msgLen, nil
  }

  return 0, nil
}

func (p1this *WebSocket) Decode(sli1msg []byte) error {
  if handshakeStatusNo == p1this.handshakeStatus {
    // 没有握手
    return p1this.p1HttpInner.Decode(sli1msg)
  } else if handshakeStatusYes == p1this.handshakeStatus {
    // 已经握手
    p1this.Sli1Msg = sli1msg
    msgLen := uint64(len(sli1msg))
    t1sli1msg := make([]byte, msgLen-uint64(p1this.headerLength))
    // 头部不需要解析，只解析数据部分
    // 解析的时候，4 个 Masking-key 轮着用
    // 第 1 个字节和第 1 个 Masking-key 异或
    // 第 2 个字节和第 2 个 Masking-key 异或
    // 第 3 个字节和第 3 个 Masking-key 异或
    // 第 4 个字节和第 4 个 Masking-key 异或
    // 第 5 个字节和第 1 个 Masking-key 异或
    var i, j uint64 = 0, uint64(p1this.headerLength)
    for j < uint64(msgLen) {
      t1sli1msg[i] = p1this.Sli1Msg[j] ^ p1this.arr1MaskingKey[i&0b00000011]
      i++
      j++
    }
    p1this.DecodeMsg = string(t1sli1msg)
  }
  return nil
}

func (p1this *WebSocket) SetDecodeMsg(msg string) {
  p1this.DecodeMsg = msg
}

func (p1this *WebSocket) Encode() ([]byte, error) {
  sli1body := []byte(p1this.DecodeMsg)
  bodyLen := len(sli1body)
  var sli1msg []byte

  if encodeTypeNoMusk == p1this.encodeType {
    if bodyLen <= 125 {
      sli1msg = make([]byte, 2)
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 0b01111111 & uint8(bodyLen)
    } else if bodyLen <= 65535 {
      sli1msg = make([]byte, 4)
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 126
      sli1msg[2] = uint8(bodyLen >> 8)
      sli1msg[3] = uint8(bodyLen >> 0)
    } else {
      sli1msg = make([]byte, 10)
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 127
      sli1msg[2] = uint8(bodyLen >> 56)
      sli1msg[3] = uint8(bodyLen >> 48)
      sli1msg[4] = uint8(bodyLen >> 40)
      sli1msg[5] = uint8(bodyLen >> 32)
      sli1msg[6] = uint8(bodyLen >> 24)
      sli1msg[7] = uint8(bodyLen >> 16)
      sli1msg[8] = uint8(bodyLen >> 8)
      sli1msg[9] = uint8(bodyLen >> 0)
    }
    sli1msg = append(sli1msg, sli1body...)

  } else {
    // 这里就直接设置 4 个 0b00000000
    var arr1maskingKey [4]byte = [4]byte{0x00, 0x00, 0x00, 0x00}
    var maskIndex int

    if bodyLen <= 125 {
      sli1msg = make([]byte, 6)
      maskIndex = 6
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 0b10000000 | uint8(bodyLen)
      sli1msg[2] = arr1maskingKey[0]
      sli1msg[3] = arr1maskingKey[1]
      sli1msg[4] = arr1maskingKey[2]
      sli1msg[5] = arr1maskingKey[3]
      sli1msg = append(sli1msg, sli1body...)
    } else if bodyLen <= 65535 {
      sli1msg = make([]byte, 8)
      maskIndex = 8
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 126
      sli1msg[2] = uint8(bodyLen >> 8)
      sli1msg[3] = uint8(bodyLen >> 0)
      sli1msg[4] = arr1maskingKey[0]
      sli1msg[5] = arr1maskingKey[1]
      sli1msg[6] = arr1maskingKey[2]
      sli1msg[7] = arr1maskingKey[3]
      sli1msg = append(sli1msg, sli1body...)
    } else {
      sli1msg = make([]byte, 14)
      maskIndex = 14
      sli1msg[0] = 0b10000000 | opcodeText
      sli1msg[1] = 127
      sli1msg[2] = uint8(bodyLen >> 56)
      sli1msg[3] = uint8(bodyLen >> 48)
      sli1msg[4] = uint8(bodyLen >> 40)
      sli1msg[5] = uint8(bodyLen >> 32)
      sli1msg[6] = uint8(bodyLen >> 24)
      sli1msg[7] = uint8(bodyLen >> 16)
      sli1msg[8] = uint8(bodyLen >> 8)
      sli1msg[9] = uint8(bodyLen >> 0)
      sli1msg[10] = arr1maskingKey[0]
      sli1msg[11] = arr1maskingKey[1]
      sli1msg[12] = arr1maskingKey[2]
      sli1msg[13] = arr1maskingKey[3]
      sli1msg = append(sli1msg, sli1body...)
    }

    var i, j = maskIndex, 0
    for j < bodyLen {
      sli1msg[i] = sli1msg[i] ^ arr1maskingKey[j&0b00000011]
      i++
      j++
    }
  }

  return sli1msg, nil
}

// SetHandshakeStatusYes 设置握手状态为已握手
func (p1this *WebSocket) SetHandshakeStatusYes() {
  p1this.handshakeStatus = handshakeStatusYes
}

// IsHandshakeStatusNo 判断是否未握手
func (p1this *WebSocket) IsHandshakeStatusNo() bool {
  return handshakeStatusNo == p1this.handshakeStatus
}

// IsHandshakeStatusYes 判断是否已握手
func (p1this *WebSocket) IsHandshakeStatusYes() bool {
  return handshakeStatusYes == p1this.handshakeStatus
}

// CheckHandshakeReq 校验握手消息（客户端申请协议升级）
func (p1this *WebSocket) CheckHandshakeReq() ([]byte, error) {
  var sli1msg []byte = []byte{}

  // 判断请求头中 connection 和 upgrade 字段是否符合要求
  connection, ok := p1this.p1HttpInner.MapHeader["connection"]
  if !ok {
    return sli1msg, errors.New("http header missing connection.")
  }
  upgrade, ok := p1this.p1HttpInner.MapHeader["upgrade"]
  if !ok {
    return sli1msg, errors.New("http header missing upgrade.")
  }
  upgradeIndex := strings.Index(connection, "Upgrade")
  websocketIndex := strings.Index(upgrade, "websocket")
  if upgradeIndex < 0 || websocketIndex < 0 {
    return sli1msg, errors.New("connection is not \"Upgrade\" or upgrade is not \"websocket\".")
  }

  // Sec-WebSocket-Accept
  secWebSocketKey, ok := p1this.p1HttpInner.MapHeader["sec-websocket-key"]
  if !ok {
    return sli1msg, errors.New("http header missing sec-webSocket-key.")
  }

  // 将 Sec-WebSocket-Key 跟 258EAFA5-E914-47DA-95CA-C5AB0DC85B11 拼接
  secAcceptStr := secWebSocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
  // 通过 SHA1 计算摘要
  secAcceptSHA1 := sha1.Sum([]byte(secAcceptStr))
  // 转成 base64 字符串
  secAcceptBase64 := base64.StdEncoding.EncodeToString(secAcceptSHA1[:])

  // 测试使用的是 JavaScript 的 WebSocket 工具
  // 在上述条件下，这几个字段都要有，少一个都跑不通
  msg := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n")
  msg += fmt.Sprintf("Connection: Upgrade\r\n")
  msg += fmt.Sprintf("Upgrade: websocket\r\n")
  msg += fmt.Sprintf("Sec-WebSocket-Accept: %s\r\n", secAcceptBase64)
  msg += fmt.Sprintf("Sec-WebSocket-Version: 13\r\n")
  msg += fmt.Sprintf("Server: tcp_server_v1\r\n\r\n")

  return []byte(msg), nil
}

// MakeHandShakeReq 构造握手消息（客户端申请协议升级）
func (p1this *WebSocket) MakeHandShakeReq() ([]byte, error) {
  t1str := "client"
  md5str := md5.Sum([]byte(t1str))
  p1this.SecWebSocketKey = base64.StdEncoding.EncodeToString(md5str[:])

  msg := fmt.Sprintf("GET /chat HTTP/1.1\r\n")
  msg += fmt.Sprintf("Upgrade: websocket\r\n")
  msg += fmt.Sprintf("Connection: Upgrade\r\n")
  msg += fmt.Sprintf("Sec-WebSocket-Key: %s\r\n", p1this.SecWebSocketKey)
  msg += fmt.Sprintf("Sec-WebSocket-Version: 13\r\n\r\n")

  return []byte(msg), nil
}

// CheckHandShakeResp 校验握手消息（服务端对客户端申请协议升级的响应）
func (this *WebSocket) CheckHandShakeResp() (err error) {
  connection, ok := this.p1HttpInner.MapHeader["connection"]
  if !ok {
    return errors.New("http header missing connection.")
  }
  upgrade, ok := this.p1HttpInner.MapHeader["upgrade"]
  if !ok {
    return errors.New("http header missing upgrade.")
  }
  upgradeIndex := strings.Index(connection, "Upgrade")
  websocketIndex := strings.Index(upgrade, "websocket")
  if upgradeIndex < 0 || websocketIndex < 0 {
    return errors.New("connection is not \"Upgrade\" or upgrade is not \"websocket\".")
  }

  secWebsocketAccept, ok := this.p1HttpInner.MapHeader["sec-websocket-accept"]
  if !ok {
    return errors.New("http header missing sec-websocket-accept.")
  }

  // 校验 sec-websocket-accept
  secAcceptStr := this.SecWebSocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
  secAcceptSHA1 := sha1.Sum([]byte(secAcceptStr))
  secAcceptBase64 := base64.StdEncoding.EncodeToString(secAcceptSHA1[:])

  if secAcceptBase64 != secWebsocketAccept {
    return errors.New("sec-websocket-accept is wrong.")
  }

  return
}
