package base

type IRequest interface {
	GetConnection() Session
	GetData() []byte
	GetMsgID() uint32
}
type Request struct {
	Ses Session
	Msg IMessage
}

func (r *Request) GetConnection() Session {
	return r.Ses
}

//GetData 获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.Msg.GetData()
}

//GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgID() uint32 {
	return r.Msg.GetMsgID()
}
