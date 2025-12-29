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

// GetAllSessionsLua возвращает список всех сессий с пользователями через Lua
func GetAllSessionsLua() ([]map[string]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis не инициализирован")
	}

	// Lua-скрипт: пройти по всем ключам и вернуть user_id и role
	script := redis.NewScript(`
		local result = {}
		local keys = redis.call('keys', '*')
		for i, key in ipairs(keys) do
			local user_id = redis.call('hget', key, 'user_id')
			local role = redis.call('hget', key, 'role')
			if user_id and role then
				table.insert(result, key .. ' => user_id:' .. user_id .. ', role:' .. role)
			end
		end
		return result
	`)

	res, err := script.Run(ctx, RedisClient, []string{}).Result()
	if err != nil {
		return nil, err
	}

	list := []map[string]string{}
	if arr, ok := res.([]interface{}); ok {
		for _, v := range arr {
			list = append(list, map[string]string{"session": fmt.Sprint(v)})
		}
	}
	return list, nil
}

// GetAllSessionsDetailed возвращает детальную информацию о всех сессиях через Lua
func GetAllSessionsDetailed() ([]map[string]interface{}, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis не инициализирован")
	}

	script := redis.NewScript(`
		local result = {}
		local keys = redis.call('keys', '*')
		
		for i, key in ipairs(keys) do
			local session_data = redis.call('hgetall', key)
			if #session_data > 0 then
				local session_info = {
					session_key = key,
					user_id = session_data[2] or 'unknown',
					role = session_data[4] or 'unknown',
					ttl = redis.call('ttl', key)
				}
				table.insert(result, session_info)
			end
		end
		
		return result
	`)

	res, err := script.Run(ctx, RedisClient, []string{}).Result()
	if err != nil {
		return nil, err
	}

	var sessions []map[string]interface{}
	if arr, ok := res.([]interface{}); ok {
		for _, v := range arr {
			if sessionMap, ok := v.([]interface{}); ok {
				session := make(map[string]interface{})
				for i := 0; i < len(sessionMap); i += 2 {
					key := fmt.Sprint(sessionMap[i])
					value := sessionMap[i+1]
					session[key] = value
				}
				sessions = append(sessions, session)
			}
		}
	}
	return sessions, nil
}
