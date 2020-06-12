package components

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

var BConfig config.Configer = nil
var ServerPort = 9091
var MaxMsgLen uint16

func init() {
	BConfig, err := config.NewConfig("ini", "config/server.conf")
	if err != nil {
		panic("config init error")
	}

	maxlines, lerr := BConfig.Int64("log::maxlines")
	if lerr != nil {
		maxlines = 1000
	}

	logConf := make(map[string]interface{})
	logConf["filename"] = BConfig.String("log::log_path")
	level, _ := BConfig.Int("log::log_level")
	logConf["level"] = level
	logConf["maxlines"] = maxlines

	confStr, err := json.Marshal(logConf)
	if err != nil {
		fmt.Println("marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(confStr))
	logs.SetLogFuncCall(true)

	ServerPort, err = BConfig.Int("server::port")
	if err != nil {
		fmt.Println("server port error is ", err)
		return
	}

	maxmsglen, err := BConfig.Int("server::max_msg_len")
	if err != nil {
		fmt.Println("server max msg len read failed , err is ", err)
		return
	}
	MaxMsgLen = uint16(maxmsglen)
}
