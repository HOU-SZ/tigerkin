package tnet

import "github.com/HOU-SZ/tigerkin/tiface"

type Request struct {
	conn tiface.IConnection //已经和客户端建立好的链接
	data []byte             //客户端请求的数据
}

// 获取请求的链接信息
func (r *Request) GetConnection() tiface.IConnection {
	return r.conn
}

// 获取请求的消息数据
func (r *Request) GetData() []byte {
	return r.data
}
