package agent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"bytes"
	"encoding/json"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/CloudCare/web-datakit-framework/log"
)

type AuthStage int8
type DataType int8

const (
	AUTH_NONE AuthStage = -1
	AUTH_INIT AuthStage = 0
	AUTH_SUCC AuthStage = -2
)

const (
	RESPONSE_DATA DataType = iota
	AUTH_DATA
)

type AuthInfo struct {
	ch    chan []byte
	stage AuthStage
}

var (
	superTopicAuth = make(map[string]AuthInfo)
)

type JsonTrans struct {
	DataType DataType    `json:"type"`
	Content  string      `json:"content"`
}

func pub_nsqd(superTopic string, data []byte) error {
	err := nsqLookupd.RandomNode().Pub(superTopic, data)
	return err
}

func handlerTopic(c *gin.Context) {
	topic := c.Query("topic")
	superTopic := c.Query("supertopic")
	if topic == "" || superTopic == "" {
		c.String(http.StatusBadRequest, "Missed topic and supertopic query parameter")
		return
	}

	key := fmt.Sprintf("%s_%s", superTopic, topic)
	status := c.Query("auth")
	if status != "" {
		//POST from python web datakit framework
		handleWebdkitAuth(c, key, topic, status)
	} else {
		//POST from third party platform: auth data or data-flow
		handleSaaS(c, key, superTopic, topic)
	}
}

func handleWebdkitAuth(c *gin.Context, key, topic, status string){
	authInfo, ok := superTopicAuth[key]

	log.Infof("topic %s status %s found %v", key, status, ok)
	defer c.Request.Body.Close()
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	switch status {
	case "0":
		if ok {
			if authInfo.stage == AUTH_SUCC {
				c.String(http.StatusCreated, "topic %s has created", topic)
			} else if authInfo.stage == AUTH_INIT {
				c.String(http.StatusBadRequest, "topic %s has inited", topic)
			} else {
				authInfo.stage = AUTH_INIT
				superTopicAuth[key] = authInfo
				c.String(http.StatusOK, "")
			}
		} else {
			auth := AuthInfo{ make(chan []byte, 1), AUTH_INIT}
			superTopicAuth[key] = auth
			c.String(http.StatusOK, "")
			go func(key string){
				select {
				case <-time.NewTicker(time.Duration(4) * time.Second).C:
					authInfo, ok := superTopicAuth[key]
					if ok &&  authInfo.stage != AUTH_SUCC {
						authInfo.stage = AUTH_NONE
						superTopicAuth[key] = authInfo
						log.Infof("topic %s delete", key)
					}
				}
			}(key)
		}
	case "-1":
		if ok {
			authInfo.stage = AUTH_NONE
			superTopicAuth[key] = authInfo
		} else {
			c.String(http.StatusNotFound, "topic %s not founed", topic)
		}
	case "-2":
		if ok {
			authInfo.stage = AUTH_SUCC
			superTopicAuth[key] = authInfo
		} else {
			c.String(http.StatusNotFound, "topic %s not founed", topic)
		}
	default:
		if ok {
			authInfo.ch <- data
			c.String(http.StatusOK, "")
		} else {
			c.String(http.StatusBadRequest, "topic %s has not exist", topic)
		}
	}
}

func handleSaaS(c *gin.Context, key, superTopic, topic string){
	var dataType DataType
	authInfo, ok := superTopicAuth[key]

	b := make([]byte, 0, 1024)
	buf := bytes.NewBuffer(b)
	c.Request.Write(buf)
	encodeString := base64.StdEncoding.EncodeToString(buf.Bytes())

	if !ok || authInfo.stage != AUTH_INIT {
		dataType = RESPONSE_DATA
	} else {
		dataType = AUTH_DATA
	}
	jsonTrans := JsonTrans{
		dataType,
		encodeString,
	}

	js, _ :=json.Marshal(jsonTrans)

	err := pub_nsqd(superTopic, js)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if ok && authInfo.stage == AUTH_INIT {
		//wait authorization data
		select {
		case auth_data := <-authInfo.ch:
			c.Status(http.StatusOK)
			c.Writer.Write(auth_data)
		case <-time.NewTicker(time.Duration(4) * time.Second).C:
			c.String(http.StatusRequestTimeout, "")
		}
	} else {
		c.String(http.StatusOK, "")
	}
}