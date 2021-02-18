package bilibili

import (
	"Walnut/pkg/app/bl"
	"bytes"
	"encoding/binary"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"net/url"
	"time"
)

// 客户端实例
type Client struct {
	RoomID      int32           // 房间 ID
	Online      int32           // 用来判断人气是否变动
	Conn        *websocket.Conn // 连接后的对象
	IsConnected bool            // 客户端是否连接
}

func (c *Client) Connect(roomId uint32) error {
	// panic("implement me")
}

// 获取一个连接好的客户端实例
func CreateClient(roomId int32) (c *Client, err error) {
	c = new(Client)

	realId, err := bl.GetRealRoomID(roomId)
	if err != nil {
		return nil, err
	}

	// 连接弹幕服务器
	u := url.URL{Scheme: "wss", Host: DanMuServer, Path: "sub"}
	c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	c.IsConnected = true
	c.RoomID = int32(realId)
	return
}

// HandShakeMsg 定义了握手包的信息格式
type HandShakeMsg struct {
	Uid       int32  `json:"uid"`
	RoomID    int32  `json:"roomid"`
	Protover  int32  `json:"protover"`
	Platform  string `json:"platform"`
	Clientver string `json:"clientver"`
	Type      int32  `json:"type"`
	Key       string `json:"key"`
}

// 返回一个初始化了的握手包信息实例
func NewHandShakeMsg(roomid int32) *HandShakeMsg {
	return &HandShakeMsg{
		Uid:       0,
		RoomID:    roomid,
		Protover:  2,
		Platform:  "web",
		Clientver: "2.4.16",
		Type:      2,
		Key:       "",
	}
}

type CMD string

var (
	RealID      = "http://api.live.bilibili.com/room/v1/Room/room_init" // params: id=xxx
	DanMuServer = "ks-live-dmcmt-bj6-pm-02.chat.bilibili.com:443"
	json        = jsoniter.ConfigCompatibleWithStandardLibrary
	P           *Pool
	UserClient  *Client

	CMDDanmuMsg                  CMD = "DANMU_MSG"                     // 普通弹幕信息
	CMDSendGift                  CMD = "SEND_GIFT"                     // 普通的礼物，不包含礼物连击
	CMDWELCOME                   CMD = "WELCOME"                       // 欢迎VIP
	CMDWelcomeGuard              CMD = "WELCOME_GUARD"                 // 欢迎房管
	CMDEntry                     CMD = "ENTRY_EFFECT"                  // 欢迎舰长等头衔
	CMDRoomRealTimeMessageUpdate CMD = "ROOM_REAL_TIME_MESSAGE_UPDATE" // 房间关注数变动
)

func NewClient() *Client {
	return &Client{
		RoomID:      0,
		Online:      0,
		Conn:        nil,
		IsConnected: false,
	}
}

// 发送握手包并开始监听消息
func (c *Client) Start(key string) (err error) {
	m := NewHandShakeMsg(c.RoomID)
	m.Key = key

	b, err := json.Marshal(m)
	if err != nil {
		return
	}

	// 发送握手包
	err = c.Send(0, 16, 1, 7, 1, b)
	if err != nil {
		return
	}

	go c.Receive()
	go c.HeartBeat([]byte("5b6f626a656374204f626a6563745d"))

	return
}

// BiliBili 客户端的 ListenerSender 接口的 Send 方法实现
// 需要发送的数据包格式如下：
// |                      首部 				    	| 实体 |
// | Len | Magic Number | Version | TypeID | Params | Data |
// 参数格式仅需要为：
// Send(magic,ver,typeID,params,data),长度通过 len(data)+16 算出
func (c *Client) Send(args ...interface{}) (err error) {
	var pLen uint32
	data := args[5].([]byte)
	if args[0] == 0 {
		pLen = uint32(len(data) + 16)
	}

	pHead := new(bytes.Buffer)

	// 首先写入计算好的包的大小
	_ = binary.Write(pHead, binary.BigEndian, pLen)
	// 依次写入首部的其他信息
	for _, val := range args[1:5] {
		if err = binary.Write(pHead, binary.BigEndian, val); err != nil {
			return
		}
	}

	// 拼接数据并写入连接对象
	sendData := append(pHead.Bytes(), data...)
	if err = c.Conn.WriteMessage(websocket.BinaryMessage, sendData); err != nil {
		return
	}

	return
}

func (c *Client) Receive() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil || msg == nil {
			c.IsConnected = false
			break
		}

		// 根据消息类型进行分类处理
		switch msg[11] {
		// 服务器发来的心跳包下行，实体部分仅直播间人气值
		case 3:
			h := ByteArrToDecimal(msg[16:])
			if int32(h) != c.Online {
				c.Online = int32(h)
				P.Online <- h
			}

		case 5:
			inflated, err := ZlibInflate(msg[16:])
			if err == nil {
				// 代表数据需要压缩，如DANMU_MSG，SEND_GIFT等信息量较大的数据包
				for len(inflated) > 0 {
					l := ByteArrToDecimal(inflated[:4])
					c := gjson.GetBytes(inflated[16:l], "cmd").String()
					switch CMD(c) {
					case CMDDanmuMsg:
						P.DanMu <- inflated[16:l]
					case CMDSendGift:
						P.Gift <- inflated[16:l]
					case CMDWELCOME:
						P.WelCome <- inflated[16:l]
					case CMDWelcomeGuard:
						P.WelComeGuard <- inflated[16:l]
					case CMDEntry:
						P.GreatSailing <- inflated[16:l]
					case CMDRoomRealTimeMessageUpdate:
						P.Fans <- inflated[16:l]
					}
					inflated = inflated[l:]
				}
			}
		}

	}
}

func (c *Client) HeartBeat(data []byte) {
	for {
		// 根据协议，每半分钟发送一次内容是 两个空对象 的数据包作为心跳包，维持连接
		if err := c.Send(0, 16, 1, 2, 1, data); err != nil {
			c.IsConnected = false
			break
		}
		time.Sleep(30 * time.Second)
	}
}
