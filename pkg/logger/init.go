package logger

import (
	"log/slog"
	"os"
)

func CreateLogger(name string) (*slog.Logger, error) {
	options := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}

	// r := &lumberjack.Logger{
	// 	Filename:   name,
	// 	LocalTime:  true,
	// 	MaxSize:    100, // megabytes
	// 	MaxAge:     28,  // days
	// 	MaxBackups: 3,
	// 	Compress:   false, // disabled by default
	// }
	// fileHandler := slog.NewTextHandler(r, options)
	// // 创建 Logger
	// logger := slog.New(fileHandler)
	// slog.SetDefault(logger)

	// 创建一个屏幕输出的 Handler
	consoleHandler := slog.NewTextHandler(os.Stdout, options)
	// 创建 Logger
	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)

	return logger, nil
}

func init() {
	CreateLogger("log_rotate.txt")
}
