package session

import (
	"time"
	"database/sql"
	"github.com/owenliang/go-crontab/lock"
	"github.com/owenliang/go-crontab/mysql"
	"github.com/owenliang/go-crontab/conf"
	"github.com/gorhill/cronexpr"
	"github.com/owenliang/go-crontab/job"
)

type Doctor struct {

}

var GDoctor *Doctor

func InitDoctor() (err error) {
	var (
		doctor Doctor
	)

	GDoctor = &doctor
	go GDoctor.checkLoop()
	return
}

// 定时检查其他会话
func (doctor *Doctor) checkLoop() {
	for {
		// 间隔一个心跳周期检查一次
		select {
		case <-time.NewTimer(time.Duration(conf.GCronConf.PingInterval) * time.Second).C:
		}

		doctor.check()
	}
}

func (doctor *Doctor) check() (err error) {
	var (
		tx *sql.Tx
		deadline int64
		sessId string
		sessRows *sql.Rows
		jobRows *sql.Rows
		sessArr []string = make([]string, 0)
		jobId int64
		expr string
		nextScheduleTime int64
	)

	if tx, err = mysql.DB.Begin(); err != nil {
		return
	}

	// 会话锁
	if err = lock.LockSession(tx); err != nil {
		goto ROLLBACK
	}

	deadline = time.Now().Unix() - int64(conf.GCronConf.KickOther)

	// 获取超时的会话
	if sessRows, err = tx.Query("SELECT `sess_id` FROM `cron_sess` WHERE `ping_time`<?", deadline); err != nil {
		goto ROLLBACK
	}
	defer sessRows.Close()

	for sessRows.Next() {
		if err = sessRows.Scan(&sessId); err != nil {
			goto ROLLBACK
		}
		// 别把自己加进去
		if sessId != GSession.sessionId {
			sessArr = append(sessArr, sessId)
		}
	}

	// 任务锁（因为用户可能并发改任务）
	if len(sessArr) != 0 {
		if err = lock.LockJob(tx); err != nil {
			goto ROLLBACK
		}
	}

	// 重置每个会话关联的任务
	for _, sessId = range sessArr {
		if jobRows, err = tx.Query("SELECT `id`,`cronexpr` FROM `cron_job` WHERE `sess_id`=?", sessId); err != nil {
			goto ROLLBACK
		}
		defer jobRows.Close()

		for jobRows.Next() {
			if err = jobRows.Scan(&jobId, &expr); err != nil {
				goto ROLLBACK
			}
			nextScheduleTime = cronexpr.MustParse(expr).Next(time.Now()).Unix()
			// 重置任务
			if _, err = tx.Exec("UPDATE `cron_job` SET `exec_id`=?,`next_schedule_time`=?,`status`=?,`sess_id`=? WHERE `id`=?", 0, nextScheduleTime, job.JOB_STATUS_WAIT, "", jobId); err != nil {
				goto ROLLBACK
			}
		}

		// 删除会话
		if _, err = tx.Exec("DELETE FROM `cron_sess` WHERE `sess_id`=?", sessId); err != nil {
			goto ROLLBACK
		}
	}

	if err = tx.Commit(); err != nil {
		goto ROLLBACK
	}
	return

ROLLBACK:
	tx.Rollback()
	return
}