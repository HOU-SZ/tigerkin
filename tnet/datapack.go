package tnet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/HOU-SZ/tigerkin/tiface"
	"github.com/HOU-SZ/tigerkin/utils"
)

// 封包拆包类，暂时不需要成员
type DataPack struct{}

// 封包拆包实例的初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//DataLen uint32(4字节) +  ID uint32(4字节)
	return 8
}

// 封包方法(压缩数据)
func (dp *DataPack) Pack(msg tiface.IMessage) ([]byte, error) {
	// 创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	// 将msgID 写进dataBuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	// 将dataLen 写进dataBuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	// 将data数据 写进dataBuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// 拆包方法(解压数据)
// 进行拆包的时候是分两次过程的，第一次得到msgId和dataLen，第二次根据dataLen读取消息数据，第二次是依赖第一次的dataLen结果，
// 所以Unpack只能解压出包头head的内容，得到msgId和dataLen。之后调用者再根据dataLen继续从io流中读取body中的数据。
func (dp *DataPack) Unpack(binaryData []byte) (tiface.IMessage, error) {
	// 创建一个存放输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	// 只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	// 读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	// 读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 判断dataLen的长度是否超出我们允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("too large msg data recieved")
	}

	// 这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
