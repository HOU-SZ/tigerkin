package tnet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 测试datapack拆包，封包功能的单元测试
func TestDataPack(t *testing.T) {
	/*
		模拟服务器
	*/
	// 创建socket TCP Server
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen error: ", err)
		return
	}

	// 创建服务器goroutine，负责从客户端goroutine读取粘包的数据，然后进行解析
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept error: ", err)
			}

			//处理客户端请求
			go func(conn net.Conn) {
				// ------ 拆包的过程 ------
				// 创建封包拆包对象dp
				dp := NewDataPack()
				for {
					// 1 第一次从conn读，读出流中的head部分
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
					if err != nil {
						fmt.Println("read head error: ", err)
						break
					}

					msgHead, err := dp.Unpack(headData) // 将headData字节流 拆包到msg中
					if err != nil {
						fmt.Println("server unpack error: ", err)
						return
					}

					// 2 第二次从conn读，根据head中的dataLen读取data内容
					if msgHead.GetDataLen() > 0 {
						// msg 是有data数据的，需要再次读取data数据
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetDataLen())

						// 根据dataLen从io中读取字节流
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data error :", err)
							return
						}

						fmt.Println("==> Recv Msg: ID=", msg.Id, ", dataLen=", msg.DataLen, ", data=", string(msg.Data))
					}
				}
			}(conn)

		}
	}()

	/*
		模拟客户端
	*/
	// 客户端goroutine，负责模拟粘包的数据，然后进行发送
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client dial err:", err)
		return
	}

	// 创建一个封包对象 dp
	dp := NewDataPack()

	// 模拟粘包过程，封装两个msg一起发送
	// 封装第一个msg包
	msg1 := &Message{
		DataLen: 5,
		Id:      0,
		Data:    []byte{'h', 'e', 'l', 'l', 'o'},
	}

	sendData1, err := dp.Pack(msg1) // 封包成二进制数据
	if err != nil {
		fmt.Println("client pack msg1 error: ", err)
		return
	}

	// 封装第二个msg包
	msg2 := &Message{
		DataLen: 8,
		Id:      1,
		Data:    []byte{'t', 'i', 'g', 'e', 'r', 'k', 'i', 'n'},
	}
	sendData2, err := dp.Pack(msg2) // 封包成二进制数据
	if err != nil {
		fmt.Println("client pack msg2 error: ", err)
		return
	}

	// 将两个sendData1和sendData2拼接一起，组成粘包
	sendData1 = append(sendData1, sendData2...)

	// 一次性发给服务器端
	conn.Write(sendData1)

	// 客户端阻塞
	select {}
}
