package lock

import (
	"database/sql"
	"sync"
)

var (
	jobLock sync.Mutex
)

func LockSession(tx *sql.Tx) (err error){
	var (
		id int64
	)

	err = tx.QueryRow("SELECT id from `cron_lock` WHERE name=? FOR UPDATE", "SESSION_LOCK").Scan(&id)
	return
}

func LockJob(tx *sql.Tx) (err error) {
	var (
		id int64
	)

	err = tx.QueryRow("SELECT id from `cron_lock` WHERE name=? FOR UPDATE", "JOB_LOCK").Scan(&id)
	return
}