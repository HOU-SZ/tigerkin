package main

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/demo_app/mmo_game/apis"
	"github.com/HOU-SZ/tigerkin/demo_app/mmo_game/core"
	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
)

// 当客户端建立连接的时候的hook函数
func OnConnecionAdd(conn tiface.IConnection) {
	// 创建一个玩家
	player := core.NewPlayer(conn)
	// 同步当前的PlayerID给客户端，发MsgID:1 消息
	player.SyncPid()
	// 同步当前玩家的初始化坐标信息给客户端，发MsgID:200消息
	player.BroadCastStartPosition()
	// 将当前新上线玩家添加到worldManager中
	core.WorldMgrObj.AddPlayer(player)
	// 将该连接绑定属性Pid
	conn.SetProperty("pid", player.Pid)
	// 告知周边玩家自己上线信息，并告知自己周边玩家信息
	player.SyncSurrounding()

	fmt.Println("=====> Player pidId = ", player.Pid, " arrived ====")
}

// 当客户端断开连接的时候的hook函数
func OnConnectionLost(conn tiface.IConnection) {
	// 获取当前连接的Pid属性
	pid, _ := conn.GetProperty("pid")

	// 根据pid获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	// 触发玩家下线业务
	if pid != nil {
		player.LostConnection()
	}

	fmt.Println("====> Player ", pid, " left =====")
}

func main() {
	// 创建服务器句柄
	s := tnet.NewServer()

	// 注册客户端连接建立和丢失函数
	s.SetOnConnStart(OnConnecionAdd)
	s.SetOnConnStop(OnConnectionLost)

	// 注册路由
	s.AddRouter(2, &apis.WorldChatApi{})
	s.AddRouter(3, &apis.MoveApi{})

	// 启动服务
	s.Serve()
}
