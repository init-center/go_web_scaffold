package mysql

import (
	"fmt"
	"go_web_scaffold/settings"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// declare a global mdb variable
var db *sqlx.DB

func Init(cfg *settings.MySQLConfig) (err error) {
	// get the db serve name from config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)
	// connect db
	// or use MustConnect to panic if the connection is unsuccessful
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		return
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	return
}

// Because db is private,
// we provide the public Close for other packages
// to close the db connection
func Close() {
	_ = db.Close()
}
