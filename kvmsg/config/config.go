package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const config_file_path = "./config.yml"

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

func GetServerAddress() (string, error) {
    c, err := LoadConfig(config_file_path)
    if err != nil {
        return "", err
    }

	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port), nil
}

func GetServerPort() (string, error) {
    c, err := LoadConfig(config_file_path)
    if err != nil {
        return "", err
    }

    return c.Server.Port, nil
}

func GetServerProtocol() (string, error) {
    c, err := LoadConfig(config_file_path)
    if err != nil {
        return "", nil
    }

    return c.Server.Protocol, nil
}

func GetClientChanBufLimit() (int, error) {
    c, err := LoadConfig(config_file_path)
    if err != nil {
        return -1, nil
    }

    return c.Client.ChannelBufferLimit, nil
}
