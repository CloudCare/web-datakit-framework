package log

import (
	"fmt"
	"log"
	"os"
)

var (
	_log *log.Logger

	isDebug bool
)

func InitLog(path string, debug bool) error {

	l, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	isDebug = debug

	_log = log.New(l, "", log.Lshortfile|log.LstdFlags)
	return nil
}

func Infof(format string, v ...interface{}) {
	_log.Output(2, fmt.Sprintf("[info ] "+format, v...))
}

func Debugf(format string, v ...interface{}) {
	if isDebug {
		_log.Output(2, fmt.Sprintf("[debug] "+format, v...))
	}
}

func Warnf(format string, v ...interface{}) {
	_log.Output(2, fmt.Sprintf("[warn] "+format, v...))
}

func Errorf(format string, v ...interface{}) {
	_log.Output(2, fmt.Sprintf("[error] "+format, v...))
}

func Fatalf(format string, v ...interface{}) {
	_log.Output(2, fmt.Sprintf("[fatal] "+format, v...))
	os.Exit(1)
}
