package tnet

import (
	"fmt"
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

//Test PreHandle
func (this *PingRouter) PreHandle(request tiface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test Handle
func (this *PingRouter) Handle(request tiface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test PostHandle
func (this *PingRouter) PostHandle(request tiface.IRequest) {
	fmt.Println("Call Router PostHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping .....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func TestServer(t *testing.T) {

	//1 创建一个server 句柄 s
	s := NewServer("[Tigerkin V0.4 test]")

	s.AddRouter(&PingRouter{})

	//2 开启服务
	go s.Serve()

	fmt.Println("Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for i := 0; i < 5; i++ {
		_, err := conn.Write([]byte("hahaha"))
		require.NoError(t, err)

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		require.NoError(t, err)
		require.NotNil(t, buf)
		// require.Equal(t, "hahaha", string(buf[:cnt]))
		fmt.Printf(" server call back : %s, cnt = %d\n", buf, cnt)

		time.Sleep(1 * time.Second)
	}
}
