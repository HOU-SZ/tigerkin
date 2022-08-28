/**
*    tigerkin server example
 */
package main

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/examples/server/router"
	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
)

// 创建连接的时候执行
func DoConnectionBegin(conn tiface.IConnection) {
	fmt.Println("DoConnecionBegin is Called ... ")

	// 设置两个链接属性，在连接创建之后
	conn.SetProperty("Name", "Hou")
	conn.SetProperty("Home", "https://github.com/HOU-SZ")
	fmt.Println("Set Connection Name, Home done!")

	err := conn.SendMsg(2, []byte("DoConnection BEGIN..."))
	if err != nil {
		fmt.Println("Connection SendMsg msg error: ", err)
	}
}

// 连接断开的时候执行
func DoConnectionLost(conn tiface.IConnection) {
	// 在连接销毁之前，查询conn的Name，Home属性
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Connection Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		fmt.Println("Connection Property Home = ", home)
	}

	fmt.Println("DoConnectionLost is Called ... ")
}

func main() {
	// 创建一个server句柄
	s := tnet.NewServer()

	// 注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	// 配置路由
	s.AddRouter(0, &router.PingRouter{})
	s.AddRouter(1, &router.HelloRouter{})

	// 开启服务
	s.Serve()
}
