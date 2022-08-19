package stream

import (
  "encoding/binary"
  "errors"
)

type Stream struct {
  // 解析状态
  ParseStatus uint8
  // 数据长度
  bodyLength uint32
  // 请求报文
  Sli1Msg []byte
  // 解析后的数据
  DecodeMsg string
}

func NewStream() *Stream {
  return &Stream{}
}

func (p1this *Stream) FirstMsgLength(sli1recv []byte) (uint64, error) {
  recvLen := uint32(len(sli1recv))
  if 0 >= recvLen {
    return 0, errors.New("STREAM_STATUS_NO_DATA")
  }
  if 4 > recvLen {
    return 0, errors.New("STREAM_STATUS_NOT_FINISH")
  }

  // 前 4 个字节，是大端字节序格式的 uint32
  p1this.bodyLength = binary.BigEndian.Uint32(sli1recv[0:4])
  if recvLen < 4+p1this.bodyLength {
    return 0, errors.New("STREAM_STATUS_NOT_FINISH")
  }

  return uint64(4 + p1this.bodyLength), nil
}

func (p1this *Stream) Decode(sli1msg []byte) error {
  p1this.DecodeMsg = string(sli1msg[4:])
  return nil
}

func (p1this *Stream) SetDecodeMsg(msg string) {
  p1this.DecodeMsg = msg
}

func (p1this *Stream) GetDecodeMsg() string {
  return p1this.DecodeMsg
}

func (p1this *Stream) Encode() ([]byte, error) {
  var sli1msg []byte

  bodyLen := len(p1this.DecodeMsg)
  if 0 >= bodyLen {
    return sli1msg, errors.New("STREAM_STATUS_NO_DATA")
  }
  t1sli1BodyLen := make([]byte, 4)
  // 把 uint32 格式的数据长度转换成大端字节序，放在最前面 4 个字节的位置上
  binary.BigEndian.PutUint32(t1sli1BodyLen, uint32(bodyLen))
  sli1msg = append(t1sli1BodyLen, p1this.DecodeMsg...)

  return sli1msg, nil
}
