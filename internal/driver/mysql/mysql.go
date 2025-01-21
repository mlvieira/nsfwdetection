package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mlvieira/nsfwdetection/internal/config"
)

const (
	maxOpenDBConn = 25
	maxIdleDBConn = 25
	maxDBLifetime = 5 * time.Minute
)

// OpenDB opens a new database connection and pings it.
func OpenDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&tls=false",
		config.AppConfig.DB.Username,
		config.AppConfig.DB.Password,
		config.AppConfig.DB.Host,
		config.AppConfig.DB.Port,
		config.AppConfig.DB.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenDBConn)
	db.SetConnMaxIdleTime(maxIdleDBConn)
	db.SetConnMaxLifetime(maxDBLifetime)

	if err = db.Ping(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}
