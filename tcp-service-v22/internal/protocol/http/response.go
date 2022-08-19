package http

import "fmt"

const (
  // 状态码
  StatusOk                  uint16 = 200
  StatusBadRequest          uint16 = 400
  StatusNotFound            uint16 = 404
  StatusInternalServerError uint16 = 500
)

var (
  // 状态码的文案
  statusText = map[uint16]string{
    StatusOk:                  "OK",
    StatusBadRequest:          "Bad Request",
    StatusNotFound:            "Not Found",
    StatusInternalServerError: "Internal Server Error",
  }
)

// Response 响应
type Response struct {
  // 状态码
  statusCode uint16
  // 响应头
  mapHeader map[string]string
}

func NewResponse() *Response {
  return &Response{
    mapHeader: make(map[string]string, 2),
  }
}

// SetHeader 设置状态码
func (p1this *Response) SetStatusCode(statusCode uint16) {
  p1this.statusCode = statusCode
}

// SetHeader 设置响应头
func (p1this *Response) SetHeader(key string, val string) {
  p1this.mapHeader[key] = val
}

// MakeMsg 构造响应报文
func (p1this *Response) MakeResponse(body string) string {
  respStr := fmt.Sprintf("HTTP/1.1 %d %v\r\n", p1this.statusCode, statusText[p1this.statusCode])

  _, ok := p1this.mapHeader["Content-Type"]
  if !ok {
    p1this.mapHeader["Content-Type"] = "text/html; charset=utf8"
  }

  for key, val := range p1this.mapHeader {
    respStr += fmt.Sprintf("%s: %s\r\n", key, val)
  }

  respStr += fmt.Sprintf("Content-Length: %v\r\n\r\n%s", len(body), body)

  return respStr
}
