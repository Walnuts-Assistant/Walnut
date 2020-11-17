package service

type Manager struct {
	Provider Provider
}

// 所有客户端实例均需要实现该接口，以具备最基本的消息收发功能
type Provider interface {
	Connect(roomId uint32) (Client, error)
	Send(args ...interface{}) error
	Receive()
	// HeartBeat(data []byte)
}

type UserDanMu struct {
	Avatar string `json:"avatar"`
	// 用户头衔
	Utitle int `json:"utitle"`
	// 用户等级
	UserLevel int `json:"user_level"`
	// 用户牌子
	MedalName string `json:"medal_name"`
	// 牌子等级
	MedalLevel int    `json:"medal_level"`
	Uname      string `json:"uname"`
	Text       string `json:"text"`
}

type UserGift struct {
	Uname  string `json:"uname"`
	Avatar string `json:"avatar"`
	Action string `json:"action"`
	Gname  string `json:"gname"`
	Nums   int32  `json:"nums"`
	Price  int    `json:"price"`
}

type WelCome struct {
	Uname string `json:"uname"`
	Title string `json:"title"`
}

type LocalInfo struct {
	MemUsedPercent float64 `json:"mem"`  // 内存使用率
	CpuUsedPercent float64 `json:"cpu"`  // CPU使用率
	SendBytes      int64   `json:"send"` // 单位时间发送字节数
}
