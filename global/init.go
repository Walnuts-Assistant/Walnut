package global

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
)

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("load config file err: %s", err)
		os.Exit(1)
	}

	Setting = new(CommonSetting)
	Setting.ServerHost = cfg.Section("Server").Key("Host").String()
}
