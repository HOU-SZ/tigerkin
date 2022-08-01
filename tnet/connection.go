package tnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/utils"
)

// 创建连接的方法
type Connection struct {
	// 当前Conn属于哪个Server
	TcpServer tiface.IServer

	// 当前连接的socket TCP套接字
	Conn *net.TCPConn

	// 当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32

	// 当前连接的关闭状态
	isClosed bool

	// // V0.2 该连接的处理方法api
	// handleAPI tiface.HandFunc

	// // V0.3 该连接的处理方法router
	// Router tiface.IRouter

	// V0.6 消息MsgId和对应业务处理api的消息管理模块
	MsgHandler tiface.IMsgHandle

	// 告知该链接已经退出/停止的channel（由Reader告知Writer退出）
	ExitBuffChan chan bool

	// 无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	// 有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	// 链接属性集合
	property map[string]interface{}
	// 保护链接属性修改的锁
	propertyLock sync.RWMutex
}

func NewConnection(server tiface.IServer, conn *net.TCPConn, connID uint32, msgHandler tiface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}

	// 将新创建的Conn添加到链接管理中
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

/*
   读消息Goroutine，用于从客户端中读取数据
*/
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), " [conn reader exit!]")
	defer c.Stop()

	for {
		// // 读取客户端的数据到buf中
		// buf := make([]byte, utils.GlobalObject.MaxPacketSize)
		// _, err := c.Conn.Read(buf)
		// if err != nil {
		// 	fmt.Println("recv buf err ", err)
		// 	c.ExitBuffChan <- true
		// 	continue
		// }

		// 创建拆包解包的对象
		dp := NewDataPack()

		// 读取客户端的Msg head（8个字节的二进制流）
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error: ", err)
			break
		}
		fmt.Printf("read headData: %+v\n", headData)

		// 拆包，得到msgId 和 dataLen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error: ", err)
			break
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error: ", err)
				break
			}
		}
		msg.SetData(data)

		// // V0.2 调用当前链接业务所绑定的handleAPI
		// if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
		// 	fmt.Println("connID ", c.ConnID, " handle is error")
		// 	c.ExitBuffChan <- true
		// 	return
		// }

		// 得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		// V0.3 从路由Routers 中找到注册绑定Conn的对应Handle
		// go func(request tiface.IRequest) {
		// 	//执行注册的路由方法
		// 	c.Router.PreHandle(request)
		// 	c.Router.Handle(request)
		// 	c.Router.PostHandle(request)
		// }(&req)

		// // V0.6 从绑定好的消息和对应的处理方法中执行对应的Handle方法
		// go c.MsgHandler.DoMsgHandler(&req)

		// V0.8 添加工作池机制，应对大量并发请求
		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经启动工作池机制，将消息交给Worker处理
			fmt.Println("Has started worker pool, send request to TaskQueue")
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			// 未启用工作池机制，从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
	c.ExitBuffChan <- true

}

/*
	写消息Goroutine，监控管道msgChan并将数据发送给客户端
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), " [conn Writer exit!]")
	defer c.Stop()

	// 不断地阻塞地等待管道msgChan的消息，一旦收到马上发给客户端
	for {
		select {
		case data, ok := <-c.msgChan:
			if ok {
				// 有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:, ", err, " Conn Writer exit")
					return
				}
				fmt.Printf("Send data success! data = %+v\n", data)
			}

		case data, ok := <-c.msgBuffChan:
			if ok {
				// 有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			}

		case <-c.ExitBuffChan:
			// conn已经关闭，代表Reader已经退出，此时Writer也应该退出
			return
		}
	}
}

//启动连接，让当前连接开始工作
func (c *Connection) Start() {
	// Start()函数结束的时候调用stop处理善后业务
	defer c.Stop()

	// 1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	// 2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()

	// 按照用户传递进来的创建连接时需要处理的业务，执行对应hook方法
	c.TcpServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			//得到退出消息，不再阻塞
			fmt.Println("Start recv ExitBuffChan...")
			return
		}
	}
}

//停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...ConnID = ", c.ConnID)
	// 如果当前链接已经关闭
	if c.isClosed {
		return
	}
	c.isClosed = true

	// 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用对应的hook方法
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()

	// 通知从缓冲队列读数据的Writer，该链接已经关闭
	c.ExitBuffChan <- true

	//将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c)

	// 关闭该链接全部管道，回收资源
	close(c.ExitBuffChan)
	close(c.msgChan)
	close(c.msgBuffChan)
}

//从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 将要发送给客户端的数据，先进行封包，再发送给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection closed when send msg")
	}
	// 将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}

	// 写进消息管道
	c.msgChan <- msg

	return nil
}

//将数据发送给缓冲队列，通过专门从缓冲队列读数据的go routine写给客户端
func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection closed when send buff msg")
	}
	// 将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 写进消息管道
	c.msgBuffChan <- msg

	return nil
}

// 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("the property was not found")
	}
}

// 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
