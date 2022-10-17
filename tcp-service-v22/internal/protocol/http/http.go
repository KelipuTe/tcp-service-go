package http

import (
	goErrors "errors"
	"strconv"
	"strings"
	"tcp-service-go/tcp-service-v22/internal/protocol"

	pkgErrors "github.com/pkg/errors"
)

const (
	ParseStatusRecvBufferEmpty uint8 = iota // 缓冲区为空
	ParseStatusNotHTTP                      // 没找到 \r\n\r\n
	ParseStatusIncomplete                   // 报文不完整（没接收全）
	ParseStatusParseErr                     // 解析出错
)

const (
	// content-type
	StrXWWWFormUrlencoded = "application/x-www-form-urlencoded"
)

var _ protocol.Protocol = &HTTP{}

// HTTP 协议
type HTTP struct {
	// ParseStatus 解析状态，详见 ParseStatus 开头的常量
	ParseStatus uint8

	// HeaderLength 请求头数据长度
	HeaderLength uint32
	// ContentLength 请求体数据长度
	ContentLength uint32

	// Sli1Msg 请求报文
	Sli1Msg []byte

	// Method 请求方法
	Method string
	// Uri 请求路由
	Uri string
	// Version 版本
	Version string

	// MapHeader 解析后的请求头
	MapHeader map[string]string
	// MapQuery 解析后的查询参数
	MapQuery map[string]string
	// MapBody 解析后的请求体
	MapBody map[string]string
}

func NewHTTP() *HTTP {
	return &HTTP{}
}

// Protocol.FirstMsgLength
func (p1this *HTTP) FirstMsgLength(sli1recv []byte) (uint64, error) {
	var firstMsgLen uint64 = 0

	recvLen := uint64(len(sli1recv))
	if 0 >= recvLen {
		p1this.ParseStatus = ParseStatusRecvBufferEmpty
		return firstMsgLen, goErrors.New("ParseStatusRecvBufferEmpty")
	}
	recvStr := string(sli1recv)

	// 找到 \r\n\r\n 的位置，用这个位置可以分隔请求头和请求体
	rnrnIndex := strings.Index(recvStr, "\r\n\r\n")
	if rnrnIndex <= 0 {
		p1this.ParseStatus = ParseStatusNotHTTP
		return firstMsgLen, goErrors.New("ParseStatusNotHTTP")
	}
	// 请求头长度等于 \r\n\r\n 的位置下标加上 \r\n\r\n 的长度
	p1this.HeaderLength = uint32(rnrnIndex + 4)

	// 找到 Content-Length 的位置
	c7l6Index := strings.Index(recvStr, "Content-Length: ")
	if c7l6Index > 0 {
		// "Content-Length: "，16 个字节
		t1recvStr := recvStr[c7l6Index+16:]
		indexR := strings.IndexByte(t1recvStr, '\r')
		// 截取 Content-Length 的字符串值
		c7l6Str := t1recvStr[0:indexR]
		c7l6Int, err := strconv.Atoi(c7l6Str)
		if nil != err {
			p1this.ParseStatus = ParseStatusParseErr
			return firstMsgLen, pkgErrors.WithMessage(err, "ParseStatusParseErr")
		}
		p1this.ContentLength = uint32(c7l6Int)
	}

	firstMsgLen = uint64(p1this.HeaderLength + p1this.ContentLength)
	if firstMsgLen > recvLen {
		// 计算出来的报文长度大于接收缓冲区中数据长度
		p1this.ParseStatus = ParseStatusIncomplete
		return firstMsgLen, goErrors.New("ParseStatusIncomplete")
	}

	return uint64(firstMsgLen), nil
}

// Protocol.Decode
func (p1this *HTTP) Decode(sli1msg []byte) error {
	p1this.Sli1Msg = sli1msg
	msg := string(p1this.Sli1Msg)
	header := msg[0:p1this.HeaderLength]
	body := msg[p1this.HeaderLength:]
	p1this.parseHeader(header)
	p1this.parseBody(body)

	return nil
}

// Protocol.Encode
func (p1this *HTTP) Encode() ([]byte, error) {
	return []byte{}, nil
}

// parseHeader 解析请求头
func (p1this *HTTP) parseHeader(header string) {
	p1this.MapHeader = make(map[string]string, 2)
	sli1header := strings.Split(header, "\r\n")
	firstLine := sli1header[0]
	// 第 1 行
	sli1firstLine := strings.Split(firstLine, " ")
	p1this.Method = sli1firstLine[0]
	p1this.Uri = sli1firstLine[1]
	p1this.Version = sli1firstLine[2]
	p1this.parseQuery(p1this.Uri)
	// 剩下的行
	for _, val := range sli1header[1:] {
		// 用 ": " 切成键值
		sli1kv := strings.Split(val, ": ")
		if 2 == len(sli1kv) {
			// 键名全部转成小写
			p1this.MapHeader[strings.ToLower(sli1kv[0])] = sli1kv[1]
		}
	}
}

// parseQuery 解析查询参数
func (p1this *HTTP) parseQuery(uri string) {
	index := strings.Index(uri, "?")
	if index > 0 {
		// 有 "?"
		p1this.Uri = uri[:index]
		query := uri[index+1:]
		if "" != query {
			// 有查询参数
			p1this.MapQuery = make(map[string]string)
			sli1query := strings.Split(query, "&")
			for _, val := range sli1query {
				arr1kv := strings.Split(val, "=")
				if 2 == len(arr1kv) {
					p1this.MapQuery[strings.ToLower(arr1kv[0])] = arr1kv[1]
				}
			}
		}
	}
}

// parseBody 解析请求体
func (p1this *HTTP) parseBody(body string) {
	c7t4, ok := p1this.MapHeader["content-type"]
	if ok {
		switch c7t4 {
		case StrXWWWFormUrlencoded:
			p1this.MapBody = make(map[string]string)
			sli1Body := strings.Split(body, "&")
			for _, val := range sli1Body {
				arr1kv := strings.Split(val, "=")
				if 2 == len(arr1kv) {
					p1this.MapBody[strings.ToLower(arr1kv[0])] = arr1kv[1]
				}
			}
		}
	}
}
