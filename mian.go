package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	loadConfig()
	server := InitWebServer()

	err := server.Run(":8081")
	if err != nil {
		panic("start server failed")
	}
}

func loadConfig() {
	configPath := pflag.String("config", "./config/dev.yaml", "config file path")
	pflag.Parse()
	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}
