package backend

import (
	"LiveAssistant/log"
	"LiveAssistant/service/bilibili"
	_ "LiveAssistant/service/bilibili"
	"github.com/go-qamel/qamel"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

func init() {
	RegisterQmlConnectFeedBack("ConnectFeedBack", 1, 0, "ConnectFeedBack")
	RegisterQmlHandleMsg("HandleMsg", 1, 0, "HandleMsg")
}

// 连接直播间模块定义
type ConnectFeedBack struct {
	qamel.QmlObject

	// _ func()       `constructor:"init"`
	_ func(int)    `slot:"receiveRoomID"`
	_ func(int)    `signal:"sendFansNums"`
	_ func(bool)   `signal:"sendConnInfo"`
	_ func(int)    `signal:"sendErr"`
	_ func(string) `signal:"sendInfo"`
}

//func (m *ConnectFeedBack) init() {
//	go func() {
//		// 3 秒一次更新一次客户端信息
//		//for {
//		//	i := GetCompInfo()
//		//	b, err := json.Marshal(i)
//		//	if err != nil {
//		//		continue
//		//	}
//		//	m.sendCompInfo(string(b))
//		//	time.Sleep(3 * time.Second)
//		//}
//	}()
//}

// receiveRoomInfo 接收选择的直播平台和房间号
// 0:BiLiBiLi 1:DouYu 2:Huya
func (m *ConnectFeedBack) receiveRoomInfo(platform,roomid int) {
	if log.CheckFileExist() {
		m.sendInfo("日志文件打开失败，可能是由于权限不足")
	}

	if c := ConnectAndServe(roomid); c == 1 {
		// 发送消息通知 QML 日志情况
		m.sendInfo("遇到错误！请查看日志文件寻找错误原因并即时告诉我们！")
	}

	// 给初次登陆的 QML 传递一个返回信息代表连接成功或失败
	//if bilibili.UserClient.IsConnected == false {
	//	m.sendErr(-1)
	//} else {
	//	m.sendErr(0)
	//}
	m.sendFansNums(GetFansByAPI(roomid))

	// 发送连接是否正常的标志
	go func() {
		for {
			if bilibili.UserClient.IsConnected == true {
				m.sendConnInfo(true)
			} else {
				m.sendConnInfo(false)
				if c := ConnectAndServe(roomid); c == 1 {
					// 发送消息通知 QML 日志情况
					m.sendInfo("遇到错误！请查看日志文件寻找错误原因并即时告诉我们！")
				}
				continue
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

// 消息处理模块定义
type HandleMsg struct {
	qamel.QmlObject
	_ func() `constructor:"init"`

	_ func(string)                 `signal:"sendDanMu"`
	_ func(string)                 `signal:"sendGift"`
	_ func(string)                 `signal:"sendWelCome"`
	_ func(string)                 `signal:"sendWelComeGuard"`
	_ func(string)                 `signal:"sendGreatSailing"`
	_ func(int)                    `signal:"sendOnlineChanged"`
	_ func(int)                    `signal:"sendFansChanged"`
	_ func(string, string, string) `signal:"sendMusicURI"` // uri,singer,name

	_      func(bool, string) `slot:"musicControl"`
	Button bool               // 点歌模块开关
	Key    string             // 点歌关键字
}

// 处理各种需要发送到 QML 的消息
func (h *HandleMsg) init() {
	go func() {
		for {
			select {
			// 处理用户弹幕
			case a := <-bilibili.P.DanMu:
				if e := GetDanMu(a); e != nil {
					s, err := json.Marshal(e)
					if err != nil {
						continue
					}
					h.sendDanMu(string(s))
					if h.Button == true {
						sp := strings.Split(e.Text, " ")
						if len(sp) > 1 && sp[0] == h.Key {
							bilibili.P.MusicInfo <- e.Text
						}
					}
				}
			// 处理用户礼物
			case b := <-bilibili.P.Gift:
				if e := GetGift(b); e != nil {
					s, err := json.Marshal(e)
					if err != nil {
						continue
					}
					h.sendGift(string(s))
				}
			// 处理贵宾进场，如老爷
			case c := <-bilibili.P.WelCome:
				if w := GetWelCome(c, 1); w != nil {
					s, err := json.Marshal(w)
					if err != nil {
						continue
					}
					h.sendWelCome(string(s))
				}
			// 处理房管进场
			case d := <-bilibili.P.WelComeGuard:
				if w := GetWelCome(d, 2); w != nil {
					s, err := json.Marshal(w)
					if err != nil {
						continue
					}
					h.sendWelComeGuard(string(s))
				}
			// 处理舰长等贵宾进场
			case e := <-bilibili.P.GreatSailing:
				if w := GetWelCome(e, 3); w != nil {
					s, err := json.Marshal(w)
					if err != nil {
						continue
					}
					h.sendGreatSailing(string(s))
				}
			// 处理关注数变动消息
			case f := <-bilibili.P.Fans:
				i := int(gjson.GetBytes(f, "data.fans").Int())
				h.sendFansChanged(i)
			// 处理在线人气变动处理
			case g := <-bilibili.P.Online:
				h.sendOnlineChanged(g)
			case j := <-bilibili.P.MusicInfo:
				s := strings.SplitN(j, " ", 2)
				uri, singer, name, err := GetMusicURI(s[1])
				if err != nil || uri == "" {
					continue
				}
				h.sendMusicURI(uri, name, singer)
			}
		}
	}()
}

// musicControl 代表客户端想要打开/关闭点歌功能
func (h *HandleMsg) musicControl(b bool, key string) {
	if b == true && key != "" {
		h.Button = true
		h.Key = key
	} else {
		h.Button = false
		h.Key = ""
	}
}
