package server

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// func main() {
// 	redisClient := newClient()
// 	result, err := ping(redisClient)
// 	if err != nil {
// 		fmt.Println(err)
// 	} else {
// 		fmt.Println(result)
// 	}
// }

func newClient() *redis.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	db, _ := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	return redisClient
}

func ping(client *redis.Client) (string, error) {
	result, err := client.Ping(client.Context()).Result()

	if err != nil {
		return "", err
	} else {
		return result, err
	}
}

func flushElasticsearch(client *redis.Client) {
	iter := client.Scan(client.Context(), 0, "log", 0).Iterator()
	for iter.Next(client.Context()) {
		client.Del(client.Context(), iter.Val())
	}
}

func flushServer(client *redis.Client, serverId string) {
	iter := client.Scan(client.Context(), 0, "*", 0).Iterator()
	for iter.Next(client.Context()) {
		response, _ := client.Get(client.Context(), iter.Val()).Result()
		if strings.Contains(response, serverId) {
			client.Del(client.Context(), iter.Val())
		}
	}
	if err := iter.Err(); err != nil {
		fmt.Println(err)
	}
}
