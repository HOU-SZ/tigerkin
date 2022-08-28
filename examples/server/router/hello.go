package router

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
)

// 自定义路由HelloRouter
type HelloRouter struct {
	tnet.BaseRouter
}

//HelloRouter Handle
func (this *HelloRouter) Handle(request tiface.IRequest) {
	fmt.Println("Call HelloRouter Handle")
	// 先读取客户端的数据，再回写
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("Hello Router"))
	if err != nil {
		fmt.Println("SendMsg error: ", err)
	}
}
