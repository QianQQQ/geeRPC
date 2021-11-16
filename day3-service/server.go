package geeRPC

import (
	"encoding/json"
	"errors"
	"fmt"
	"geeRPC/codec"
	"geeRPC/service"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Register(obj interface{}) error {
	svc := service.NewService(obj)
	if _, dup := s.serviceMap.LoadOrStore(svc.Name, svc); dup {
		return errors.New("rpc: service already defined: " + svc.Name)
	}
	return nil
}

func Register(obj interface{}) error { return DefaultServer.Register(obj) }

func (s *Server) findService(serviceMethod string) (svc *service.Service, mType *service.MethodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svcInterface, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svcInterface.(*service.Service)
	mType = svc.Methods[methodName]
	fmt.Println(svc.Methods)
	if mType == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return svc, mType, nil
}

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer conn.Close()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for {
		// 读取请求
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			// 不用处理请求, 直接回复请求
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		// 处理请求 + 回复请求
		go s.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	cc.Close()
}

// 我不理解
func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	// 获取请求头信息
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	// 构造请求
	req := &request{h: h}
	req.service, req.methodType, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv, req.replyv = req.methodType.NewArgv(), req.methodType.NewReplyv()
	// ReadBody的参数是指针
	argvInterface := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvInterface = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvInterface); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	err := req.service.Call(req.methodType, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		s.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

// 需要锁, 因为发送是互斥的(就像打印)
func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}
