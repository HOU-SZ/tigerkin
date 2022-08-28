/**
*    tigerkin client example
 */
package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/HOU-SZ/tigerkin/tnet"
)

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start error, exit!")
		return
	}

	for {
		// 封包并发送message消息
		dp := tnet.NewDataPack()
		msg, _ := dp.Pack(tnet.NewMsgPackage(0, []byte("Tigerkin client example test MsgID=0, [Ping]")))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("write error: ", err)
			return
		}

		// 先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) // ReadFull会把headData填充满为止
		if err != nil {
			fmt.Println("read head error: ", err)
			break
		}

		// 将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack error: ", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			// msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*tnet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			// 根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data error: ", err)
				return
			}

			fmt.Println("==> Test Router:[Ping] Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}

		time.Sleep(1 * time.Second)
	}
}
