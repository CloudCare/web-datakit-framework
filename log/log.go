package log

import (
	"log"
	"os"
)

var (
	_log *log.Logger

	isDebug bool
)

func InitLog(path string, debugModel bool) error {

	l, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	isDebug = debugModel

	_log = log.New(l, "", log.Lshortfile|log.LstdFlags)
	return nil
}

func Infof(format string, v ...interface{}) {
	_log.Printf("[info ] "+format, v...)
}

func Debugf(format string, v ...interface{}) {
	if isDebug {
		_log.Printf("[debug] "+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	_log.Printf("[warn ] "+format, v...)
}

func Errorf(format string, v ...interface{}) {
	_log.Printf("[error] "+format, v...)
}

func Fatalf(format string, v ...interface{}) {
	_log.Printf("[fatal] "+format, v...)
	os.Exit(1)
}
