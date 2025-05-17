package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	APIs   []string `mapstructure:"apis"`
	Google struct {
		ListId string `mapstructure:"list-id"`
	} `mapstructure:"google"`
}

func InitConfig(cfgPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config into struct: %v", err)
	}

	return cfg, nil
}
