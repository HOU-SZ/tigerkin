package tiface

import "net"

//定义连接接口
type IConnection interface {

	// 启动连接，让当前连接开始工作
	Start()

	// 停止连接，结束当前连接状态
	Stop()

	// 从当前连接获取原始的socket TCPConn
	GetTCPConnection() *net.TCPConn

	// 获取当前连接ID
	GetConnID() uint32

	// 获取远程客户端地址信息
	RemoteAddr() net.Addr

	// 将数据发送给无缓冲队列，通过专门从队列读数据的goroutine写给TCP客户端（无缓冲）
	SendMsg(msgId uint32, data []byte) error

	// 将数据发送给有缓冲队列，通过专门从缓冲队列读数据的goroutine写给TCP客户端（有缓冲）
	SendBuffMsg(msgId uint32, data []byte) error
}

// //定义一个统一处理链接业务的接口
// type HandFunc func(*net.TCPConn, []byte, int) error
