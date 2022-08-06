package apis

import (
	"fmt"

	"github.com/HOU-SZ/tigerkin/demo_app/mmo_game/core"
	"github.com/HOU-SZ/tigerkin/demo_app/mmo_game/pb"
	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/tnet"
	"google.golang.org/protobuf/proto"
)

type WorldChatApi struct {
	tnet.BaseRouter
}

func (*WorldChatApi) Handle(request tiface.IRequest) {
	// 1. 将客户端传来的proto协议解码
	msg := &pb.Talk{}
	err := proto.Unmarshal(request.GetData(), msg)
	if err != nil {
		fmt.Println("Talk Unmarshal error ", err)
		return
	}

	// 2. 得知当前的消息是从哪个玩家传递来的,从连接属性pid中获取
	pid, err := request.GetConnection().GetProperty("pid")
	if err != nil {
		fmt.Println("GetProperty pid error", err)
		request.GetConnection().Stop()
		return
	}
	// 3. 根据pid得到player对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	// 4. 让player对象发起聊天广播请求
	player.Talk(msg.Content)
}
