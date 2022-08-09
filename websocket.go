package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Shitomo/play-kafka-chat-core/adapter/gateway"
	"github.com/Shitomo/play-kafka-chat-core/driver/logger"
	"github.com/Shopify/sarama"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//チェックオリジンにはリクエスト元のアドレスをチェックする関数をセットする．
	//今回の場合は常にtrueを返す関数をセットしている=どんなアドレスからのリクエストも許容する
	CheckOrigin: func(r *http.Request) bool { return true },
}

func websocketHandler(users *Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
		}
		// IDはとりあえず固定
		newUser := NewUser(UserID(1), conn)
		users.Append(newUser)
	}
}

func start(ctx context.Context, consumer sarama.PartitionConsumer, users *Users) {
	gateway := gateway.NewRealtimeMessageConsumer(consumer)
	go func() {
		ch, errCh := gateway.Consume(ctx)
		for {
			select {
			case message := <-ch:
				if message.SenderId == "" {
					continue
				}
				logger.Infof(ctx, "%v", message)
				err := users.Send(UserID(1), message)
				if err == nil {
					continue
				}
				// 該当ユーザーが未接続の場合、スキップする
				if errors.As(err, &UserNotFoundError{}) {
					continue
				}
				logger.Errorf(ctx, "error while sending message %v to user %d. caused by %v", message, UserID(1), err)
			case err := <-errCh:
				if err != nil {
					logger.Errorf(ctx, "error while waiting. caused %v", err)
				}
			}
		}
	}()
}
