package util

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Redis *Redis `yaml:"redis"`
}

type Redis struct {
	Network              string `yaml:"network"`
	Address              string `yaml:"address"`
	Password             string `yaml:"password"`
	DefaultTopic         string `yaml:"defaultTopic"`
	DefaultConsumerGroup string `yaml:"defaultConsumerGroup"`
	DefaultConsumerID    string `yaml:"defaultConsumerID"`
}

func (cfg *Config) LoadConfig(path string) (err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	config := Config{}
	err = viper.Unmarshal(&config)
	cfg.Redis = config.Redis
	fmt.Println(cfg.Redis.Network)
	return err
}

type ConfigInterface interface {
	LoadConfig(path string) (err error)
}
