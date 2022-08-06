package tnet

import (
	"errors"
	"fmt"
	"net"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/utils"
)

var (
	tigerkinLogo = `                                        
		😝 👻 🐯  𝓣𝓲𝓰𝓮𝓻𝓴𝓲𝓷  🐯 👻 😝 
                                        `
	topLine    = `┌─────────────────────────────────────────────────────────────┐`
	borderLine = `│         [Github] https://github.com/HOU-SZ/tigerkin         │`
	bottomLine = `└─────────────────────────────────────────────────────────────┘`
)

//iServer 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgId和对应的业务处理api
	msgHandler tiface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr tiface.IConnManager
	// 该Server的连接创建时Hook函数
	OnConnStart func(conn tiface.IConnection)
	// 该Server的连接断开时的Hook函数
	OnConnStop func(conn tiface.IConnection)
}

//============== 定义当前客户端链接的handle api ===========
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	//回显业务
	fmt.Println("[Conn Handle] CallBackToClient ... ")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err ", err)
		return errors.New("CallBackToClient error")
	}
	return nil
}

//============== 实现 tiface.IServer 里的全部接口方法 ========

//开启网络服务
func (s *Server) Start() {
	fmt.Printf("[Tigerkin] Server name: %s, listen at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Tigerkin] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	//开启一个go routine去做服务端Listener业务
	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("Resolve tcp addr err: ", err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("Listen", s.IPVersion, "err", err)
			return
		}

		// 已经监听成功
		fmt.Println("[Tigerkin] Start Tigerkin server [", s.Name, "] success, now listenning...")

		// TODO 应该有一个自动生成ID的方法
		var cid uint32
		cid = 0

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}
			fmt.Println("Get client connection, remote address = ", conn.RemoteAddr().String())

			//3.2 设置服务器最大连接控制,如果超过最大连接包，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				// TODO 给客户端响应一个超出最大连接数的错误包
				conn.Close()
				continue
			}

			//3.3 将处理该连接请求的业务方法（此处为CallBackToClient，回显业务）和conn绑定，得到Connection对象
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++

			//3.4 启动当前链接的处理业务
			go dealConn.Start()
		}
	}()
}

// 停止服务
func (s *Server) Stop() {
	fmt.Println("[Tigerkin] Tigerkin server, name ", s.Name, "has STOPED!")

	// 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

// 运行服务
func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}
}

//路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router tiface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}

// 得到当前server的链接管理模块
func (s *Server) GetConnMgr() tiface.IConnManager {
	return s.ConnMgr
}

// 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(tiface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(tiface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn tiface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("------Call onConnStart()------")
		s.OnConnStart(conn)
	}
}

// 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn tiface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("------Call onConnStop()------")
		s.OnConnStop(conn)
	}
}

/*
  创建一个服务器句柄
*/
func NewServer() tiface.IServer {
	printLogo()
	// 由于import utils包时，会自动执行init操作，因此无需下面操作
	// //先初始化全局配置文件
	// utils.GlobalObject.Reload()

	s := &Server{
		Name:       utils.GlobalObject.Name, //从全局参数GlobalObject获取
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,    //从全局参数GlobalObject获取
		Port:       utils.GlobalObject.TcpPort, //从全局参数GlobalObject获取
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}

	return s
}

func printLogo() {
	fmt.Println(tigerkinLogo)
	fmt.Println(topLine)
	fmt.Println(borderLine)
	fmt.Println(bottomLine)
}
