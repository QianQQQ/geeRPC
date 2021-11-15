package geeRPC

import "geeRPC/codec"

const MagicNumber = 0x3bef5c

// 协商编解码方式
type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}
