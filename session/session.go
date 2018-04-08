package session

import (
	"github.com/owenliang/go-crontab/mysql"
	"github.com/owenliang/go-crontab/conf"
	"time"
	"fmt"
	"database/sql"
	"github.com/owenliang/go-crontab/lock"
	"os"
)

// 定时ping
type Session struct {
	sessionId string
	pingTime int64
}

var GSession *Session

func InitSession() (err error) {
	var (
		session Session
		tx *sql.Tx
	)

	session.sessionId = conf.GCronConf.NodeName + "#" + fmt.Sprintf("%d", time.Now().UnixNano() / 1000 / 1000)

	if tx, err = mysql.DB.Begin(); err != nil {
		return
	}

	session.pingTime = time.Now().Unix()

	if _, err = tx.Exec("INSERT INTO `cron_sess`(`sess_id`,`ping_time`) VALUES(?,?)", session.sessionId, session.pingTime); err != nil {
		goto ROLLBACK
	}

	if err = tx.Commit(); err != nil {
		goto ROLLBACK
	}

	GSession = &session

	go GSession.pingLoop()
	return

ROLLBACK:
	tx.Rollback()
	return
}

func (session *Session) pingLoop() {
	var (
		now int64
		err error
		killMyself bool = false
	)
	for {
		select {
			case <- time.NewTimer(1 * time.Second).C:
		}

		now = time.Now().Unix()
		if now - session.pingTime >= int64(conf.GCronConf.PingInterval) {
			if killMyself, err = session.ping(); err == nil {
				session.pingTime = time.Now().Unix()
			}
		}

		// 自杀
		if killMyself ||  now - session.pingTime >= int64(conf.GCronConf.KillMyself) {
			os.Exit(2)
		}
	}
}

func (session *Session) ping() (killMyself bool, err error) {
	var (
		tx *sql.Tx
		id int64
	)

	killMyself = false

	if tx, err = mysql.DB.Begin(); err != nil {
		return
	}

	// 会话锁
	if err = lock.LockSession(tx); err != nil {
		goto ROLLBACK
	}

	// 查询会话记录
	if err = tx.QueryRow("SELECT id FROM `cron_sess` WHERE `sess_id`=?", session.sessionId).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			killMyself = true
		}
		goto ROLLBACK
	}

	// 更新心跳时间
	if _, err = tx.Exec("UPDATE `cron_sess` SET `ping_time`=? WHERE `sess_id`=?", session.pingTime, session.sessionId); err != nil {
		goto ROLLBACK
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		goto ROLLBACK
	}
	return

ROLLBACK:
	tx.Rollback()
	return
}