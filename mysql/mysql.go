package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/owenliang/go-crontab/conf"
)

var DB *sql.DB

// 初始化mysql连接池
func InitMysql() (err error) {
	var (
		db *sql.DB
	)

	if db, err = sql.Open("mysql", conf.GCronConf.Dsn); err != nil {
		return
	}
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	DB = db

	return
}