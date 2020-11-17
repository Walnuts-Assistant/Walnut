package global

var (
	//Global struct
	Setting *CommonSetting

	//BarrageServerListsURI 获取BiliBili弹幕服务器列表，连接时必要的token
	//*****BiliBili******
	//params: room_id=xxx&platform=pc&player=web
	BarrageServerListsURI = "https://api.live.bilibili.com/room/v1/Danmu/getConf"

	Providers = make(map[string])
)