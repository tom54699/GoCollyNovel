package redis

import (
	"context"

	"github.com/GoCollyNovel/go-novel-scraper/configs"
	"github.com/GoCollyNovel/go-novel-scraper/pkg/logger"
	"github.com/go-redis/redis/v8"
)

var Rd *redis.Client

func ConnectRedis() {
	Rd = redis.NewClient(&redis.Options{
		Addr:     configs.C.Redis.Addr,
		Password: configs.C.Redis.Password,
		DB:       configs.C.Redis.DB, // 使用默認的資料庫
	})

	_, err := Rd.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("無法連接到Redis: ", err)
		return
	}
}

func CloseRedis() {
	defer Rd.Close()
}
