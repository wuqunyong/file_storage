package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func CreateLogger(name string) (*slog.Logger, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v, filename:%s", err, file)
	}

	options := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}

	// fileHandler := slog.NewTextHandler(file, options)

	// 创建一个屏幕输出的 Handler
	consoleHandler := slog.NewTextHandler(os.Stdout, options)

	// 创建 Logger
	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)

	return logger, nil
}
