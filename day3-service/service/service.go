package service

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type Service struct {
	Self    reflect.Value          // 结构体实例本身
	Name    string                 // 结构体的名称
	Typ     reflect.Type           // 结构体的类型
	Methods map[string]*MethodType // 所包含的方法
}

// 这里的反射, 人麻了
func NewService(self interface{}) *Service {
	s := &Service{
		Self: reflect.ValueOf(self),
	}
	s.Name = reflect.Indirect(s.Self).Type().Name()
	s.Typ = reflect.TypeOf(self)
	if !ast.IsExported(s.Name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.Name)
	}
	s.registerMethods()
	return s
}

func (s *Service) registerMethods() {
	s.Methods = map[string]*MethodType{}
	for i := 0; i < s.Typ.NumMethod(); i++ {
		method := s.Typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 || mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.Methods[method.Name] = &MethodType{
			Method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.Name, method.Name)
	}
}

func (s *Service) Call(m *MethodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.Method.Func
	returnValues := f.Call([]reflect.Value{s.Self, argv, replyv})
	if errInterface := returnValues[0].Interface(); errInterface != nil {
		return errInterface.(error)
	}
	return nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
