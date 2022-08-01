package main

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/demo_app/mmo_game/core"
	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
)

// 当客户端建立连接的时候的hook函数
func OnConnecionAdd(conn tiface.IConnection) {
	temp_pid := 1
	// 创建一个玩家
	player := core.NewPlayer(conn, int32(temp_pid))
	// 同步当前的PlayerID给客户端，发MsgID:1 消息
	player.SyncPid()
	// 同步当前玩家的初始化坐标信息给客户端，发MsgID:200消息
	player.BroadCastStartPosition()

	fmt.Println("=====> Player pidId = ", temp_pid, " arrived ====")
}

func main() {
	// 创建服务器句柄
	s := tnet.NewServer("MMO Game Tigerkin")

	// 注册客户端连接建立和丢失函数
	s.SetOnConnStart(OnConnecionAdd)

	// 启动服务
	s.Serve()
}
