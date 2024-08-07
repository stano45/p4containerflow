package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <redis-server-ip>")
		return
	}

	redisIP := os.Args[1]
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:6379", redisIP),
	})

	ctx := context.Background()

	for {
		val, err := rdb.Get(ctx, "counter").Result()
		if err == redis.Nil {
			fmt.Println("Key 'counter' does not exist. Setting to 1.")
			if err := rdb.Set(ctx, "counter", 1, 0).Err(); err != nil {
				log.Fatalf("Failed to set 'counter': %v", err)
			}
		} else if err != nil {
			log.Fatalf("Failed to get 'counter': %v", err)
		} else {
			counter, err := strconv.Atoi(val)
			if err != nil {
				log.Fatalf("Failed to convert 'counter' to int: %v", err)
			}

			counter++
			if err := rdb.Set(ctx, "counter", counter, 0).Err(); err != nil {
				log.Fatalf("Failed to set 'counter': %v", err)
			}
			fmt.Printf("counter = %d\n", counter)
		}

		time.Sleep(1 * time.Second)
	}
}
