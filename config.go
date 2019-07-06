package main

import "github.com/spf13/viper"

var (
	config Config
)

func init() {
	viper.SetConfigFile("config.yml")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config file: " + err.Error())
	}
	if err := viper.Unmarshal(&config); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}
}

type Config struct {
	Xigua    Xigua
	Postgres Postgres
}

type Xigua struct {
	Account  string
	Password string
}

type Postgres struct {
	DBName     string
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     int
}
