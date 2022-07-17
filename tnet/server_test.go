package tnet

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/stretchr/testify/require"
)

//ping test 自定义路由
type PingRouter struct {
	BaseRouter
}

//Test Handle
func (this *PingRouter) Handle(request tiface.IRequest) {
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

func TestServer(t *testing.T) {

	//1 创建一个server 句柄 s
	s := NewServer("[Tigerkin V0.5 test]")

	s.AddRouter(&PingRouter{})

	//2 开启服务
	go s.Serve()

	fmt.Println("Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		// _, err := conn.Write([]byte("hahaha"))
		// require.NoError(t, err)
		// buf := make([]byte, 512)
		// cnt, err := conn.Read(buf)
		// require.NoError(t, err)
		// require.NotNil(t, buf)
		// // require.Equal(t, "hahaha", string(buf[:cnt]))
		// fmt.Printf(" server call back : %s, cnt = %d\n", buf, cnt)

		// 发送封包message消息
		dp := NewDataPack()
		msg, _ := dp.Pack(NewMsgPackage(0, []byte("ping")))
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
			require.Equal(t, "pong", string(msg.Data))
		}

		time.Sleep(1 * time.Second)
	}
}
