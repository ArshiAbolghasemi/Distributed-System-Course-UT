package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Protocol string `mapstructure:"protocol"`
	} `mapstructure:"server"`
	Client struct {
		ChannelBufferLimit int `mapstructure:"channelBufferLimit"`
	} `mapstructure:"client"`
}

func LoadConfig(filename string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(filename)
	err := v.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("config: %v", err)
	}

	var c Config
	err = v.Unmarshal(&c)
	if err != nil {
		return Config{}, fmt.Errorf("config: %v", err)
	}

	return c, nil
}

func (c Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}
