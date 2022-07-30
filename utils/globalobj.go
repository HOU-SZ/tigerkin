package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/HOU-SZ/tigerkin/tiface"
	log "github.com/sirupsen/logrus"
)

/*
	存储一切有关Tigerkin框架的全局参数，供其他模块使用
	一些参数是可以通过tigerkin.json由用户进行配置
*/

type GlobalObj struct {
	/*
		Server
	*/
	TcpServer tiface.IServer //当前的全局Server对象
	Host      string         //当前服务器主机IP
	TcpPort   int            //当前服务器主机监听端口号
	Name      string         //当前服务器名称

	/*
		Tigerkin
	*/
	Version       string //当前Tigerkin版本号
	MaxPacketSize uint32 //当前框架数据包的最大值
	MaxConn       int    //当前服务器主机允许的最大链接个数

	WorkerPoolSize   uint32 //业务工作Worker池的goroutine数量
	MaxWorkerTaskLen uint32 //每个worker对应的消息队列中任务数量的最大值

	MaxMsgChanLen uint32 //SendBuffMsg发送消息的缓冲最大长度
}

/*
	定义一个全局的对象
*/
var GlobalObject *GlobalObj

//读取用户的配置文件
func (g *GlobalObj) Reload() {

	data, err := ioutil.ReadFile("conf/tigerkin.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("The file tigerkin.json doesn't exist, will use default config.")
			return
		} else {
			panic(err)
		}
	}
	//将json数据解析到struct中
	//fmt.Printf("json :%s\n", data)
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

/*
	提供init方法，默认加载，初始化当前的GlobalObject
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:          "TigerkinServerApp",
		Version:       "V0.11",
		TcpPort:       7777,
		Host:          "0.0.0.0",
		MaxConn:       100,
		MaxPacketSize: 4096,

		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
	}

	//从配置文件conf/tigerkin.json中加载一些用户配置的参数
	GlobalObject.Reload()
}
