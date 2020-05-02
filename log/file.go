package log

import (
	"os"
	"time"
)

var (
	LogSavePath = "runtime/logs/"
	TimeFormat  = "20060102"
)

func openLogFile() *os.File {
	_, err := os.Stat(LogSavePath)
	switch {
	case os.IsNotExist(err):
		dir, _ := os.Getwd()
		err := os.MkdirAll(dir+"/"+LogSavePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	case os.IsPermission(err):
		return nil
	}

	f, err := os.OpenFile(LogSavePath+time.Now().Format(TimeFormat)+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	return f
}
