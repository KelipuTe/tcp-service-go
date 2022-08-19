package protocol

// 协议名称
const (
  TCPStr       string = "tcp"
  HTTPStr      string = "http"
  StreamStr    string = "stream"
  WebSocketStr string = "websocket"
)

// mapsupported 支持的协议
var mapsupported map[string]bool

func init() {
  mapsupported = map[string]bool{
    TCPStr:       true,
    HTTPStr:      true,
    StreamStr:    true,
    WebSocketStr: true,
  }
}

// IsSupported 判断协议是否支持
func IsSupported(name string) bool {
  _, ok := mapsupported[name]
  return ok
}

const (
  StrStream = "stream"
)

// Protocol 协议
type Protocol interface {
  // FirstMsgLength 计算接收缓冲区中第 1 个完整的报文的长度
  FirstMsgLength(sli1recv []byte) (uint64, error)
  // Decode 报文解码
  Decode(sli1msg []byte) error
  // Encode 报文编码
  Encode() ([]byte, error)
}
