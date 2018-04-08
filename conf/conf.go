package conf

import (
	"encoding/json"
	"io/ioutil"
)

type CronConf struct{
	Dsn string // mysql地址
	NodeName string // 节点名
	PingInterval int // 心跳间隔
	KillMyself int // 超过killMysel秒心跳失败则自杀
	KickOther int // 剔除其他超过kickOther秒没有心跳的节点
}

var GCronConf *CronConf

// 加载配置
func LoadCronConf(filename string) (err error) {
	var (
		content []byte
		cronConf CronConf
	)

	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = json.Unmarshal(content, &cronConf); err != nil {
		return
	}

	GCronConf = &cronConf

	return
}