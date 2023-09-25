package util

import "github.com/spf13/viper"

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

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return config, err
}
