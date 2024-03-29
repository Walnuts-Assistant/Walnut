package backend

import (
	//"LiveAssistant/log"
	"LiveAssistant/service/bilibili"
	"Walnut/pkg/app/bl"
	"Walnut/pkg/log"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var
(
	json        = jsoniter.ConfigCompatibleWithStandardLibrary
	userInfoUrl = "https://api.bilibili.com/x/space/acc/info"           // mid=382297465&jsonp=jsonp
	RoomInfoURI = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom" // params:?room_id=923833
)

const (
	// TODO 舰长身份的识别
	CommonUser = 0      // 普通用户
	Vip        = 1 << 1 // 老爷
	Guard      = 1 << 2 // 房管
	Sailing    = 1 << 3 // 大航海
)

// ConnectAndServe 重新维持客户端连接
// 1:日志正常打印，通知用户查看日志
// 2:创建日志文件错误，权限不足，打开文件失败等
// 3.
func ConnectAndServe(roomid int) (code int) {
	key, err := bl.GetAccessKey(int32(roomid))
	if err != nil {
		//log.Info("获取连接弹幕服务器所需的 key 失败")
		return 1
	}

	// 获取客户端实例
	bilibili.UserClient, err = bilibili.CreateClient(int32(roomid))
	if err != nil || bilibili.UserClient == nil {
		log.Info("客户端创建失败")
		return 1
	}

	// 启动客户端
	err = bilibili.UserClient.Start(key)
	if err != nil {
		bilibili.UserClient.IsConnected = false
		log.Info("客户端启动失败")
		return 1
	}
	return 0
}



// GetUserAvatar 获取用户头像
func GetUserAvatar(userid int32) (ava string, err error) {
	u := fmt.Sprintf("%s?mid=%d&jsonp=jsonp", userInfoUrl, userid)

	resp, err := http.Get(u)
	if err != nil {
		return
	}

	rawdata, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()
	if err != nil {
		return
	}
	ava = gjson.GetBytes(rawdata, "data.face").String()

	return
}

// GetDanMu 提取一条弹幕
func GetDanMu(src []byte) *UserDanMu {
	d := new(UserDanMu)
	u := json.Get(src, "info", 2, 0).ToInt32()
	a, err := GetUserAvatar(u)
	if err != nil {
		return nil
	}
	d.Avatar = a

	d.Uname = gjson.GetBytes(src, "info.2.1").String()
	d.Text = gjson.GetBytes(src, "info.1").String()
	d.MedalName = gjson.GetBytes(src, "info.3.1").String()
	d.MedalLevel = int(gjson.GetBytes(src, "info.3.0").Int())
	d.UserLevel = int(gjson.GetBytes(src, "info.4.0").Int())
	// 判定用户称呼，比如 房管 | 老爷 | 舰长等等，用二进制位按位与表示

	guard := int(gjson.GetBytes(src, "info.2.2").Int())
	vip := int(gjson.GetBytes(src, "info.2.3").Int())
	d.Utitle = guard<<1 | vip<<2 | CommonUser

	return d
}

// GetGift 获取一条礼物信息
func GetGift(src []byte) *UserGift {
	g := new(UserGift)

	g.Uname = gjson.GetBytes(src, "data.uname").String()
	g.Avatar = gjson.GetBytes(src, "data.face").String()
	g.Action = gjson.GetBytes(src, "data.action").String()
	g.Gname = gjson.GetBytes(src, "data.giftName").String()
	g.Nums = int32(gjson.GetBytes(src, "data.num").Int())
	g.Price = int(gjson.GetBytes(src, "data.price").Int())

	if g.Price == 0 {
		return nil
	}
	return g
}

// typeID = 1是老爷，2是房管，3是舰长/提督等大航海
func GetWelCome(src []byte, typeID uint8) *WelCome {
	w := new(WelCome)
	var s string
	switch typeID {
	case 1:
		w.Uname = gjson.GetBytes(src, "data.uname").String()
		level := int(gjson.GetBytes(src, "data.svip").Int())
		if level == 1 {
			w.Title = "年费老爷"
		} else {
			w.Title = "老爷"
		}
	case 2:
		w.Uname = gjson.GetBytes(src, "data.username").String()
		w.Title = "房管"
	case 3:
		s = gjson.GetBytes(src, "data.copy_writing").String()
		b := []byte(s)
		w.Uname = string(b[15 : len(b)-18])
		w.Title = string(b[6:13])
	}
	if w.Uname == "" || w.Title == "" {
		return nil
	}
	return w
}

// 根据歌手名和歌曲获取音乐URI地址
func GetMusicURI(keywords string) (URI, singer, name string, err error) {
	// 根据歌手名，音乐名获取歌曲id
	q := url.Values{}
	q.Set("keywords", keywords)
	q.Set("limit", "1")
	u := url.URL{
		Scheme:   "http",
		Host:     server,
		Path:     "/search",
		RawQuery: q.Encode(),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}

	rawdata, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	id := int(gjson.GetBytes(rawdata, "result.songs.0.id").Int())
	name = gjson.GetBytes(rawdata, "result.songs.0.name").String()
	singer = gjson.GetBytes(rawdata, "result.songs.0.artists.0.name").String()

	// 根据id获取歌曲uri
	r := fmt.Sprintf("http://%s/song/url?id=%d", server, id)
	res, err := http.Get(r)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(res.Body)

	_ = resp.Body.Close()
	if err != nil {
		return
	}

	URI = gjson.GetBytes(data, "data.0.url").String()

	return
}

// 根据官方的api获取关注数
func GetFansByAPI(roomid int) int {
	u := fmt.Sprintf("%s?room_id=%d", RoomInfoURI, roomid)

	resp, err := http.Get(u)
	if err != nil {
		return 0
	}

	rawdata, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()
	if err != nil {
		return 0
	}
	fans := int(gjson.GetBytes(rawdata, "data.anchor_info.relation_info.attention").Int())

	return fans
}

// 获取客户端CPU MEM 网络信息
func GetCompInfo() (l *LocalInfo) {
	l = new(LocalInfo)
	vm, _ := mem.VirtualMemory()
	f, _ := cpu.Percent(time.Second, false)
	f[0], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", f[0]), 64)
	io, _ := net.IOCounters(true)
	for _, v := range io {
		// qamel 不支持uint64类型，转换一下
		l.SendBytes += int64(v.BytesSent)
	}

	// 不判错，若获取失败返回零值
	l.MemUsedPercent = vm.UsedPercent
	l.CpuUsedPercent = f[0]
	l.SendBytes = l.SendBytes / 1024

	//TODO 磁盘使用率，读写量暂定

	return
}
