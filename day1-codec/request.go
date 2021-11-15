package geeRPC

import (
	"geeRPC/codec"
	"reflect"
)

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}
