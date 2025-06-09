package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	APIs   map[string]bool `mapstructure:"apis"`
	Google struct {
		Sync   bool   `mapstructure:"sync"`
		ListId string `mapstructure:"list-id"`
	} `mapstructure:"google"`
	SyncOnStartup bool `mapstructure:"sync-on-startup"`

	MaxLength struct {
		Title       int `mapstructure:"title"`
		Description int `mapstructure:"description"`
	} `mapstructure:"max-length"`
}

func LoadConfig(cfgPath string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

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

func SetGoogleSync(v bool) {
	viper.Set("google.sync", v)
	viper.WriteConfig()
}

func SetSyncOnStartup(v bool) {
	viper.Set("sync-on-startup", v)
	viper.WriteConfig()
}

func SetAPIs(apis map[string]bool) {
	viper.Set("apis", apis)
	viper.WriteConfig()
}
