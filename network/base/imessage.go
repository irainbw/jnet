package base

//消息封装
type IMessage interface {
	GetDataLen() uint32 //获取消息数据段长度
	SetMsgID(uint32)    //消息ID
	SetData([]byte)     //消息内容
	GetMsgID() uint32   //获取消息ID
	GetData() []byte    //获取消息内容
}

const (
	SessionConnect uint32 = iota + 1
	SessionClose
)

type Message struct {
	DataLen uint32 //消息的长度
	ID      uint32 //消息的ID
	Data    []byte //消息的内容
}

func NewMsgPackage(ID uint32, data []byte) *Message {
	return &Message{
		DataLen: uint32(len(data)),
		ID:      ID,
		Data:    data,
	}
}

func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

func (m *Message) SetMsgID(id uint32) {
	m.ID = id
}

func (m *Message) SetData(data []byte) {
	m.Data = data
}

func (m *Message) GetMsgID() uint32 {
	return m.ID
}
func (m *Message) GetData() []byte {
	return m.Data
}
