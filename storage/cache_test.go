package storage

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func BenchmarkGetListFullV1(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		b.Fatal(err)
	}

	opt := fmt.Sprintf("user=usertest password=%s dbname=usertest sslmode=disable",
		os.Getenv("CACHE_DB_PASSWD"))
	db, err := sqlx.Connect("postgres", opt)
	if err != nil {
		b.Fatal(err)
	}

	var bananas []Banana
	if err := GetList(rdb, db, BananaDataSource, &bananas); err != nil {
		b.Fatal(err)
	}

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		if err := GetListFullV1(rdb, db, BananaDataSource, &bananas); err != nil {
			b.Fatalf("Error in GetListFullV1: %v", err)
		}
	}

	err = rdb.Close()
	if err != nil {
		b.Fatal("Error closing connection:", err)
		return
	}
}

func BenchmarkGetListFullV2(b *testing.B) {
	internmap := make(StellarCache)

	opt := fmt.Sprintf("user=usertest password=%s dbname=usertest sslmode=disable",
		os.Getenv("CACHE_DB_PASSWD"))
	db, err := sqlx.Connect("postgres", opt)
	if err != nil {
		b.Fatal(err)
	}

	var bananas []Banana
	if err := GetListFullV2(internmap, db, BananaDataSource, &bananas); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if err := GetListFullV2(internmap, db, BananaDataSource, &bananas); err != nil {
			b.Fatalf("Error in GetListFullV2: %v", err)
		}
	}
}
