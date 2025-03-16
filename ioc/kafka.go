package ioc

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webok/internal/events"
	"webok/internal/events/article"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	fmt.Printf("kafka addr: %v\n", cfg.Addr)
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		// 这里可以使用日志库
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
