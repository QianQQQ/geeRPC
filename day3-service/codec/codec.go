package codec

import "io"

type Header struct {
	ServiceMethod string // Service.Method
	Seq           uint64 // 请求的序号, 区分不同的请求
	Error         string // 服务端如果发生错误, 置于Error中
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const GobType Type = "application/gob"
const JsonType Type = "application/json"

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = map[Type]NewCodecFunc{}
	NewCodecFuncMap[GobType] = NewGobCodec
}
