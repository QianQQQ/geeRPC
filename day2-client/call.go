package geeRPC

// 一次RPC调用所需要的信息
// 应该是client发给server的东西
type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

func (c *Call) done() {
	c.Done <- c
}
