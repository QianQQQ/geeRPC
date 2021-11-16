package geeRPC

import (
	"geeRPC/codec"
	"geeRPC/service"
	"reflect"
)

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	service      *service.Service
	methodType   *service.MethodType
}
