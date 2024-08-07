package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"github.com/go-redis/redis/v8"
	"context"
	"encoding/json"
)

type Server struct {
	RedisClient *redis.Client
}

var ctx = context.Background()

func (s *Server) getData(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is missing", http.StatusBadRequest)
		return
	}

	val, err := s.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching data from Redis", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"key":   key,
		"value": val,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	redisAddr := flag.String("redisAddr", "localhost:6379", "Address of the Redis server")
	flag.Parse()

	rdb := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	server := &Server{RedisClient: rdb}

	http.HandleFunc("/get", server.getData)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
