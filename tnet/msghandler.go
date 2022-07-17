package tnet

import (
	"fmt"
	"strconv"

	"github.com/HOU-SZ/tigerkin/tiface"
)

/*
	消息处理模块的实现
*/
type MsgHandle struct {
	Apis map[uint32]tiface.IRouter //存放每个MsgId 所对应的处理方法
}

// 创建MsgHandle的方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis: make(map[uint32]tiface.IRouter),
	}
}

// 马上以非阻塞方式处理消息，调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request tiface.IRequest) {
	// 根据MsgID找到对应的Router
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is NOT FOUND!")
		return
	}

	// 执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router tiface.IRouter) {
	// 1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	// 2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId, " success!")
}
