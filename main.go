package main

import (
	"github.com/owenliang/go-crontab/conf"
	"flag"
	"runtime"
	"fmt"
	"os"
	"github.com/owenliang/go-crontab/mysql"
	"github.com/owenliang/go-crontab/session"
	"time"
)

var (
	config string
)

func initCmd() {
	flag.StringVar(&config, "config", "./go-crontab.json", "path to go-crontab.json")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	// 初始化程序环境
	initEnv()

	// 解析命令行
	initCmd()

	// 加载配置
	if err = conf.LoadCronConf(config); err != nil {
		goto ERROR
	}

	// 初始化数据库连接池
	if err = mysql.InitMysql(); err != nil {
		goto ERROR
	}

	// 初始化session心跳线程
	if err = session.InitSession(); err != nil {
		goto ERROR
	}

	// 初始化session健康检查
	if err = session.InitDoctor(); err != nil {
		goto ERROR
	}

	for {
		time.Sleep(time.Second * 1)
	}
	return

ERROR:
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
