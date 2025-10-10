package handler

import (
	"log/slog"

	"github.com/Ben1524/GoMall/common/elk_log"
)

// 统一错误处理
var errLogger, _ = elk_log.NewDefaultLogger("localhost:5000")

func ErrorHandle(err error) {
	if err != nil {
		errLogger.Error("server internal error", slog.String("error", err.Error()))
	}
}
