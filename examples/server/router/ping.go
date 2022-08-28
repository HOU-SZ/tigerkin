package router

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
)

// 自定义路由PingRouter
type PingRouter struct {
	tnet.BaseRouter
}

// PingRouter Handle
func (this *PingRouter) Handle(request tiface.IRequest) {

	fmt.Println("Call PingRouter Handle")
	// 先读取客户端的数据，再回写
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(0, []byte("pong"))
	if err != nil {
		fmt.Println("SendMsg error: ", err)
	}
}
