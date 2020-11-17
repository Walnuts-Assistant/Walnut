package bl

import (
	"Walnut/global"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

//GetAccessKey 是获取发送握手包必须的 key
func GetAccessKey(roomJd int32) (key string, err error) {
	u := fmt.Sprintf("%s?room_id=%d&platform=pc&player=web", global.BarrageServerListsURI, roomJd)

	resp, err := http.Get(u)
	if err != nil {
		return
	}

	rawdata, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()
	if err != nil {
		return
	}
	key = gjson.GetBytes(rawdata, "data.token").String()

	return
}

func GetRealRoomID(short int32) (realID int, err error) {
	u := fmt.Sprintf("%s?id=%d", RealID, short)
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("http.Get token err: ", err)
		return 0, err
	}

	rawdata, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Println("ioutil.ReadAll(resp.Body) err: ", err)
		return 0, err
	}
	realID = int(gjson.GetBytes(rawdata, "data.room_id").Int())
	return realID, nil
}