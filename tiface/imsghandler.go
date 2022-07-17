package tiface

/*
	消息管理抽象层
*/
type IMsgHandle interface {
	DoMsgHandler(request IRequest)          // 马上以非阻塞方式处理消息，调度/执行对应的Router消息处理方法
	AddRouter(msgId uint32, router IRouter) // 为消息添加具体的处理逻辑
}
