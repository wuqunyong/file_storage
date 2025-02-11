package logger

import (
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

func CreateLogger(name string) (*slog.Logger, error) {
	r := &lumberjack.Logger{
		Filename:   name,
		LocalTime:  true,
		MaxSize:    100, // megabytes
		MaxAge:     28,  // days
		MaxBackups: 3,
		Compress:   false, // disabled by default
	}

	options := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}

	fileHandler := slog.NewTextHandler(r, options)

	// 创建一个屏幕输出的 Handler
	// consoleHandler := slog.NewTextHandler(os.Stdout, options)

	// 创建 Logger
	logger := slog.New(fileHandler)
	slog.SetDefault(logger)

	return logger, nil
}

func init() {
	CreateLogger("log_rotate.txt")
}
