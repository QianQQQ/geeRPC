package geeRPC

import (
	"geeRPC/codec"
	"time"
)

const MagicNumber = 0x3bef5c

// 协商编解码方式
// 使用json来编解码
type Option struct {
	MagicNumber       int
	CodecType         codec.Type
	ConnectionTimeout time.Duration
	HandleTimeout     time.Duration
}

var DefaultOption = &Option{
	MagicNumber:       MagicNumber,
	CodecType:         codec.GobType,
	ConnectionTimeout: time.Second * 10,
}
