package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"example/cache/storage"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln(err)
		return
	}

	opt := fmt.Sprintf("user=usertest password=%s dbname=usertest sslmode=disable",
		os.Getenv("CACHE_DB_PASSWD"))
	db, err := sqlx.Connect("postgres", opt)
	if err != nil {
		log.Fatalln(err)
	}

	var bananas []storage.Banana
	if err := storage.GetList(rdb, db, storage.BananaDataSource, &bananas); err != nil {
		log.Fatalln(err)
	}
	log.Println(bananas)

	err = rdb.Close()
	if err != nil {
		log.Println("Error closing connection:", err)
		return
	}
}
