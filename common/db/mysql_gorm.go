package db

import (
	"fmt"
	"log/slog"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func NewMysqlGorm(cfg *config.Config) (db *gorm.DB, err error) {
	err = nil
	db = nil
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	slog.Info("mysql dsn", slog.String("dsn", dsn))
	//"user:password@tcp(
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		slog.Error("mysql connect err", slog.String("dsn", dsn))
		panic(err)
	}
	return
}
