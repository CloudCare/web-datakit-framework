package agent

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	"github.com/CloudCare/web-datakit-framework/conf"
	"github.com/CloudCare/web-datakit-framework/log"
	"github.com/CloudCare/web-datakit-framework/nsq"

	"github.com/gin-gonic/gin"
)

var (
	routeVerifyBash = make(map[string]string)

	nsqLookupd *nsq.NSQLookupd
)

func LoadNSQLookupd() error {
	nsqLookupd = nsq.New(conf.Cfg.Global.NSQAddr)
	if err := nsqLookupd.Refresh(); err != nil {
		log.Errorf("nsqlookupd refresh failed, %s", err.Error())
		return err
	}

	log.Infof("NSQLookupd find %d NSQ nodes", len(nsqLookupd.Producers))
	return nil
}

func Server(addr string) {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	for _, v := range conf.Cfg.Callbacks {
		routeVerifyBash[v.Route] = v.Bash

		// route string prefix not '/'
		router.POST("/"+v.Route, func(c *gin.Context) { handlerCallback(c) })
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Infof("listen ip:port %s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf(err.Error())
	}
}

func handlerCallback(c *gin.Context) {

	var err error
	var body []byte

	if verifyBash := routeVerifyBash[c.Request.URL.Path[1:]]; verifyBash != "" {
		out, _err := exec.Command("bash", verifyBash).Output()
		if _err != nil {
			err = errors.New("bash script exec failed, " + _err.Error())
			goto __END__
		}

		c.String(http.StatusOK, string(out))
		routeVerifyBash[c.Request.URL.Path] = ""
		return
	}

	body, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		goto __END__
	}
	c.Request.Body.Close()

	if len(body) == 0 {
		err = errors.New("invalid message body size 0")
		goto __END__
	}

	err = nsqLookupd.RandomNode().Pub(c.Request.URL.Path[1:], body)
	if err != nil {
		goto __END__
	}

	log.Debugf("callback process success，url: %s, body: %s", c.Request.URL.Path, string(body))
	c.String(http.StatusOK, "OK")
	return

__END__:

	log.Errorf("callback process failed，url: %s, err: %s", c.Request.URL.Path, err.Error())
	c.String(http.StatusBadRequest, err.Error())
}

func timerCycle() {
	ticker := time.NewTicker(time.Second * conf.Cfg.Global.TimerCycle)
	defer ticker.Stop()
	content := []byte{0}

	for {
		select {
		case <-ticker.C:
			if err := nsqLookupd.RandomNode().Pub(conf.Cfg.Global.TimerTopic, content); err != nil {
				log.Errorf("timer send nsq failed, %s", err.Error())
			}
		}
	}

}
