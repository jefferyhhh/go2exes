package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// New 创建日志实例
func New(level, logFile string) (*logrus.Logger, error) {
	logger := logrus.New()

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %w", err)
		}
		logger.SetOutput(file)
	}

	return logger, nil
}
