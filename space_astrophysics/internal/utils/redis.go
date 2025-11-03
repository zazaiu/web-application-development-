package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var RedisClient *redis.Client

func InitRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// SetSession сохраняет токен в Redis
func SetSession(token string, userID int, role string, ttl time.Duration) error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.HSet(ctx, token, "user_id", userID, "role", role).Err()
}

// GetSession получает токен из Redis
func GetSession(token string) (int, string, error) {
	if RedisClient == nil {
		return 0, "", nil
	}
	data, err := RedisClient.HGetAll(ctx, token).Result()
	if err != nil || len(data) == 0 {
		return 0, "", err
	}
	userID := 0
	role := data["role"]
	fmt.Sscanf(data["user_id"], "%d", &userID)
	return userID, role, nil
}

// DeleteSession удаляет токен из Redis
func DeleteSession(token string) error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Del(ctx, token).Err()
}
