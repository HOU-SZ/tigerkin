package tnet

import (
	"fmt"
	"strconv"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/utils"
)

/*
	消息处理模块的实现
*/
type MsgHandle struct {
	// 存放每个MsgId 所对应的处理方法
	Apis map[uint32]tiface.IRouter
	// 业务工作Worker池的worker数量
	WorkerPoolSize uint32
	// Worker取任务的消息队列
	TaskQueue []chan tiface.IRequest
}

// 创建MsgHandle的方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]tiface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,                               //从全局配置中获取
		TaskQueue:      make([]chan tiface.IRequest, utils.GlobalObject.WorkerPoolSize), // 一个worker对应一个queue
	}
}

// 将消息交给TaskQueue， 由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request tiface.IRequest) {
	// 根据ConnID来分配当前的request应该由哪个worker负责处理
	// 采用轮询的平均分配法则，保证每个worker所收到的request任务是均衡的
	// 由哪个worker处理，把这个request发送给对应的TaskQueue即可
	// TODO 目前只考虑单体应用，轮询分配，优化：分布式场景，优化分配方式，考虑区域，借鉴envoy负载均衡策略

	// 得到需要处理此request的ConnID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	// fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(), " request msgID = ", request.GetMsgID(), "to workerID = ", workerID)
	// 将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
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
	// fmt.Println("[Tigerkin] Add api msgId = ", msgId, " success!")
}

// 启动worker工作池（只执行一次，因为一个框架只能有一个工作池）
func (mh *MsgHandle) StartWorkerPool() {
	// 根据WorkerPoolSize依次开启worker，每个worker为一个goroutine
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个worker被启动
		// 给当前worker在对应的channel任务队列开辟空间，第i个worker就用第i个channel
		mh.TaskQueue[i] = make(chan tiface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		// 启动当前Worker，阻塞等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个worker
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan tiface.IRequest) {
	fmt.Println("[Tigerkin] Worker ID = ", workerID, " has started.")
	// 不断的等待队列中的消息
	for {
		select {
		// 有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}
