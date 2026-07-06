package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Init 初始化配置
func Init(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		exePath, err := os.Executable()
		if err == nil {
			viper.AddConfigPath(filepath.Dir(exePath))
		}
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "")
	viper.SetDefault("excel.default_sheet", "Sheet1")

	viper.SetEnvPrefix("TOOL")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置失败: %w", err)
		}
	}
	return nil
}

func GetString(key string) string { return viper.GetString(key) }
func GetInt(key string) int       { return viper.GetInt(key) }
func GetBool(key string) bool     { return viper.GetBool(key) }
