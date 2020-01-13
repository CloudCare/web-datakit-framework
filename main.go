package main

import (
	"flag"
	"fmt"

	"github.com/CloudCare/web-datakit-framework/agent"
	"github.com/CloudCare/web-datakit-framework/conf"
	"github.com/CloudCare/web-datakit-framework/log"
)

const _VERSION = "v1.0.2"

var (
	flagConfig  = flag.String("cfg", "wdf.conf", "configure file")
	flagVersion = flag.Bool("version", false, "print version")
)

func main() {

	flag.Parse()

	if *flagVersion {
		fmt.Println("WDF Version: ", _VERSION)
		return
	}

	if err := conf.LoadConfig(*flagConfig); err != nil {
		fmt.Println("load config failed, ", err.Error())
		return
	}

	if err := log.InitLog(conf.Cfg.Global.LogPath, conf.Cfg.Global.LogDebug); err != nil {
		fmt.Println("initiation log option failed ", err.Error())
		return
	}

	if err := agent.LoadNSQLookupd(); err != nil {
		fmt.Println("load nsq lookup failed,  ", err.Error())
		return
	}

	agent.Server("0.0.0.0" + conf.Cfg.Global.Listen)
}
