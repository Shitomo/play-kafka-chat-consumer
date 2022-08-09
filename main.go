package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/Shitomo/play-kafka-chat-core/adapter/gateway"
	"github.com/Shitomo/play-kafka-chat-core/driver/logger"
	"github.com/Shopify/sarama"
)

var (
	// kafkaのアドレス
	bootstrapServers = "localhost:9092"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	users := NewUsers()

	http.HandleFunc("/", websocketHandler(&users))

	flag.Parse()

	if bootstrapServers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	logger.Infof(ctx, "broker url is %s", bootstrapServers)

	brokers := strings.Split(bootstrapServers, ",")
	config := sarama.NewConfig()

	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Fatal(ctx, err)
		}
	}()

	partition, err := consumer.ConsumePartition(gateway.RealtimeMessageTopic, 0, sarama.OffsetNewest)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	go start(ctx, partition, &users)

	http.ListenAndServe(":8082", nil)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<-signals
}
