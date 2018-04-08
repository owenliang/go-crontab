package main

import (
	"github.com/gorhill/cronexpr"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"fmt"
	"database/sql"
	"os"
	"context"
)

func main()  {
	// cronexpr
	now := time.Now()
	triggerTime := cronexpr.MustParse("* * * *  *").Next(now)
	fmt.Println(now, triggerTime)

	// mysql
	db, err := sql.Open("mysql", "root:baidu@123@tcp(localhost:3306)/go-crontab")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	// timeoutCtx, _ := context.WithTimeout(context.TODO(), time.Duration(1 * time.Second))
	timeoutCtx := context.TODO()
	tx, _ := db.BeginTx(timeoutCtx, nil)
	stmt, _ := tx.PrepareContext(timeoutCtx, "SELECT * FROM `cron_lock` WHERE name=? FOR UPDATE")
	rows, err := stmt.QueryContext(timeoutCtx, "JOB_LOCK")

	defer rows.Close()
	defer stmt.Close()

	for rows.Next() {
		var id int64
		var name string
		rows.Scan(&id, &name)
		fmt.Println( id, name)
	}

	for true {
		time.Sleep(time.Duration(time.Second * 1))
	}
}
