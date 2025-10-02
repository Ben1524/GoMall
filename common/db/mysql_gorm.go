package db

import (
	"fmt"
	"github.com/Ben1524/GoMall/common/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log/slog"
	"strconv"
)

func NewMysqlGorm(cfg *config.Config) (db *gorm.DB, err error) {
	err = nil
	db = nil
	port, err := strconv.Atoi(cfg.Database.Port) // 为了兼容postgresql，端口号用字符串表示
	if err != nil {
		slog.Error("mysql port atoi err", slog.String("port", cfg.Database.Port))
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		port,
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
