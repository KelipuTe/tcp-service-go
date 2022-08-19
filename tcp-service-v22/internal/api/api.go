package api

const (
  TypeRequest  uint8 = iota // 请求
  TypeResponse              // 响应
)

const (
  // gateway 接收服务提供者的注册信息
  ActionRegisteServiceProvider string = "registe_service_provider"

  // gateway 对服务提供者的心跳检测
  ActionPing string = "ping"
  ActionPong string = "pong"
)

// 自定义的交互数据包
type APIPackage struct {
  // 数据包的 ID
  // 一般来说是 TCP 连接的 IP，用于区分这个数据包是谁的
  Id string
  // Type 数据包类型，详见 Type 开头的常量
  Type uint8
  // Action 访问的方法，详见 Action 开头的常量
  Action string
  // 数据（经过 json 格式化的结构体）
  Data string
}

// ReqInRegisteServiceProvider，ActionRegisteServiceProvider 对应的数据结构
type ReqInRegisteServiceProvider struct {
  Name      string   `json:"name"`
  Sli1Route []string `json:"route"`
}
