package tnet

import (
	"errors"
	"fmt"
	"net"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/utils"
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
	fmt.Printf("[Tigerkin] Server name: %s,listen at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Tigerkin] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	//开启一个go routine去做服务端Listener业务
	go func() {
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
		fmt.Println("Start Tigerkin server  ", s.Name, " success, now listenning...")

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

			//3.2 TODO Server.Start() 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接

			//3.3 将处理该连接请求的业务方法（此处为CallBackToClient，回显业务）和conn绑定，得到Connection对象
			dealConn := NewConnection(conn, cid, s.msgHandler)
			cid++

			//3.4 启动当前链接的处理业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Tigerkin server , name ", s.Name)

	//TODO  Server.Stop() 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
}

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

/*
  创建一个服务器句柄
*/
func NewServer(name string) tiface.IServer {
	// 由于import utils包时，会自动执行init操作，因此无需下面操作
	// //先初始化全局配置文件
	// utils.GlobalObject.Reload()

	s := &Server{
		Name:       utils.GlobalObject.Name, //从全局参数GlobalObject获取
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,    //从全局参数GlobalObject获取
		Port:       utils.GlobalObject.TcpPort, //从全局参数GlobalObject获取
		msgHandler: NewMsgHandle(),
	}

	return s
}
