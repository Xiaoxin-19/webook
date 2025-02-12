package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"log"
)

func main() {
	loadConfig()
	initLogger()
	server := InitWebServer()

	err := server.Run(":8081")
	if err != nil {
		panic("start server failed")
	}
}

func loadConfig() {
	err := viper.AddRemoteProvider("etcd3", "http://localhost:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
		return
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Printf("config changed!!!!\n")
	})
	err = viper.ReadRemoteConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
