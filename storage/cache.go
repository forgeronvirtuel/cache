package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type DataSource struct {
	Key   string
	Query string
}

var (
	BananaDataSource = DataSource{Key: "datastore/Bananas", Query: "SELECT * FROM Banana"}
	AppleDataSource  = DataSource{Key: "datastore/Apples", Query: "SELECT * FROM Apple"}
)

type Banana struct {
	Name string `db:"name" json:"name"`
}

type Apple struct {
	Name string `db:"name" json:"name"`
}

func CreateBanana(db *sqlx.DB, item *Banana) error {
	_, err := db.Exec(`INSERT INTO banana (name) VALUES ($1)`, item.Name)
	if err != nil {
		return errors.Wrap(err, "create error")
	}
	return nil
}

func getJSONValueFromRedis(rdb *redis.Client, key string, value interface{}) error {
	entityJSON, err := rdb.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return err
	}

	if err != nil {
		return errors.Wrap(err, "Trying to get data from redis")
	}

	err = json.Unmarshal(entityJSON, value)
	if err != nil {
		return errors.Wrap(err, "Trying to unmarshal data from redis")
	}

	return nil
}

func setJSONValueIntoRedis(rdb *redis.Client, key string, value interface{}) error {
	valueIntoJSON, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "Trying to marshal data from db into JSON")
	}

	err = rdb.Set(context.Background(), key, valueIntoJSON, time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "Trying to set data into Redis")
	}
	return nil
}

func redisDecorator(rdb *redis.Client, key string, f func(interface{}) error, values interface{}) error {
	// step 1: Check the redis cache for data
	// 		   If data found, stop here.
	err := getJSONValueFromRedis(rdb, key, values)
	if err == nil {
		return nil
	}
	if err != redis.Nil {
		return errors.Wrap(err, "While retrieving entities")
	}

	// step 2: Check the source of truth
	if err := f(values); err != nil {
		return err
	}

	// step 3: Update the cache
	setJSONValueIntoRedis(rdb, key, values)

	return nil
}

func GetList(rdb *redis.Client, db *sqlx.DB, ds DataSource, values interface{}) error {
	f := func(value interface{}) error {
		if err := sqlx.Select(db, value, ds.Query); err != nil {
			return errors.Wrap(err, "Trying to get data from db")
		}
		return nil
	}

	if err := redisDecorator(rdb, ds.Key, f, values); err != nil {
		return err
	}

	return nil
}

func GetListFull(rdb *redis.Client, db *sqlx.DB, ds DataSource, values interface{}) error {
	// step 1: Check the redis cache for data
	// 		   If data found, stop here.
	err := getJSONValueFromRedis(rdb, ds.Key, values)
	if err == nil {
		return nil
	}
	if err != redis.Nil {
		return errors.Wrap(err, "While retrieving entities")
	}

	// step 2: Check the source of truth
	if err := sqlx.Select(db, values, ds.Query); err != nil {
		return errors.Wrap(err, "Trying to get data from db")
	}

	// step 3: Update the cache
	setJSONValueIntoRedis(rdb, ds.Key, values)

	return nil
}
