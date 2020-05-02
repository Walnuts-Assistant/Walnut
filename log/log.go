package log

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

var (
	logger    *log.Logger
	logPrefix string
)

func ChechFileExist() bool {
	f := openLogFile()
	if f != nil {
		logger = log.New(f, "", log.LstdFlags)
		return true
	} else {
		return false
	}
}

func Info(v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		logPrefix = fmt.Sprintf("[%s:%d]", filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[runtime.Caller err!]")
	}

	logger.SetPrefix(logPrefix)
	logger.Println(v)
}
