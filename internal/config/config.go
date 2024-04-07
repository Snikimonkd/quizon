package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ReadConfigFile - прочитать конфиг
func ReadConfigFile() error {
	viper.SetConfigName("local_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".config")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("can't read config: %w", err)
	}

	return nil
}
