package api

const (
  APIUserName  string = "/api/user_name"
  APIUserLevel string = "/api/user_level"
)

// ReqInUserName APIUserName 对应的数据结构
type ReqInUserName struct {
  Id   uint64 `json:"id"`
  Name string `json:"name"`
}

// ReqInUserLevel APIUserLevel 对应的数据结构
type ReqInUserLevel struct {
  Id    uint64 `json:"id"`
  Level uint8  `json:"level"`
}
