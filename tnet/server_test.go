package tnet

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/stretchr/testify/require"
)

//ping test 自定义路由
type PingRouter struct {
	BaseRouter
}

type HelloRouter struct {
	BaseRouter
}

// PingRouter Handle
func (router *PingRouter) Handle(request tiface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	// _, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	// if err != nil {
	// 	fmt.Println("call back ping ping ping error")
	// }
	// 先读取并验证客户端的数据，再回复客户端
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	// 回写数据
	err := request.GetConnection().SendMsg(1, []byte("pong"))
	if err != nil {
		fmt.Println(err)
	}
}

// HelloRouter Handle
func (router *HelloRouter) Handle(request tiface.IRequest) {
	fmt.Println("Call HelloRouter Handle")
	// 先读取并验证客户端的数据，再回复客户端
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("Hello Tigerkin"))
	if err != nil {
		fmt.Println(err)
	}
}

func DoConnectionBegin(conn tiface.IConnection) {
	fmt.Println("=============DoConnectionBegin is called=============")
	conn.SetProperty("Name", "Shizheng Hou")
	conn.SetProperty("GitHub", "https://github.com/HOU-SZ")
}

func DoConnectionEnd(conn tiface.IConnection) {
	fmt.Println("=============DoConnectionEnd is called=============")

	if name, err := conn.GetProperty("name"); err != nil {
		fmt.Println("Name = ", name)
	}
	if github, err := conn.GetProperty("GitHub"); err != nil {
		fmt.Println("GitHub = ", github)
	}
}

func StartTestServer() {
	//1 创建一个server 句柄 s
	s := NewServer()

	//2 注册当前链接的hook函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionEnd)

	//3 注册路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloRouter{})

	//4 开启服务
	go s.Serve()

	fmt.Println("Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)
}

func TestServer(t *testing.T) {

	StartTestServer()

	var wg sync.WaitGroup

	for i := 0; i < int(2); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", "127.0.0.1:7777")
			require.NoError(t, err)

			testMap := make(map[int][2]string)
			testMap[0] = [2]string{"ping", "pong"}
			testMap[1] = [2]string{"hello", "Hello Tigerkin"}

			for key, value := range testMap {
				// 发送封包message消息
				dp := NewDataPack()
				msg, _ := dp.Pack(NewMsgPackage(uint32(key), []byte(value[0])))
				_, err := conn.Write(msg)
				require.NoError(t, err)

				// 先读出流中的head部分
				headData := make([]byte, dp.GetHeadLen())
				_, err = io.ReadFull(conn, headData) // ReadFull 会把msg填充满为止
				require.NoError(t, err)

				// 将headData字节流 拆包到msg中
				msgHead, err := dp.Unpack(headData)
				require.NoError(t, err)

				if msgHead.GetDataLen() > 0 {
					// msg 是有data数据的，需要读取data数据
					msg := msgHead.(*Message)
					msg.Data = make([]byte, msg.GetDataLen())

					// 根据dataLen从io中读取字节流
					_, err := io.ReadFull(conn, msg.Data)
					require.NoError(t, err)

					fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
					require.Equal(t, uint32(1), msg.Id)
					require.Equal(t, value[1], string(msg.Data))
				}

				time.Sleep(1 * time.Second)
			}
		}()
	}
	wg.Wait()
}
