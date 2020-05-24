package service

// 所有客户端实例均需要实现该接口，以具备最基本的消息收发功能
type ListenerSender interface {
	Send(args ...interface{}) error
	Receive()
	HeartBeat(data []byte)
}
