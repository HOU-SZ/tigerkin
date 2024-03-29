package tnet

type Message struct {
	// 消息的长度
	DataLen uint32
	// 消息的ID
	Id uint32
	// 消息的内容
	Data []byte
}

// 创建一个Message消息包
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		DataLen: uint32(len(data)),
		Id:      id,
		Data:    data,
	}
}

// 获取消息数据段长度
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

// 获取消息ID
func (msg *Message) GetMsgId() uint32 {
	return msg.Id
}

// 获取消息内容
func (msg *Message) GetData() []byte {
	return msg.Data
}

// 设置消息数据段长度
func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

// 设置消息ID
func (msg *Message) SetMsgId(msgId uint32) {
	msg.Id = msgId
}

// 设置消息内容
func (msg *Message) SetData(data []byte) {
	msg.Data = data
}
