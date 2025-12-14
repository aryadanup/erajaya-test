package app

import (
	"context"
	"erajaya-test/shared/datastore"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Database struct {
	Postgres *gorm.DB
	Redis    *redis.Client
}

func InitDatabase(ctx context.Context) *Database {

	db := &Database{}

	db.Postgres = initPostgres(ctx)
	db.Redis = initRedis(ctx)

	return db
}

func (db *Database) Close(ctx context.Context) {

	if db.Postgres != nil {

		sqlDB, err := db.Postgres.DB()
		if err != nil {
			log.Printf("error connecting to Postgres: %v", err)
		} else {
			if err := sqlDB.Close(); err != nil {
				log.Printf("error closing Postgres: %v", err)
			} else {
				log.Println("postgres connection closed")
			}
		}
	}

	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			log.Printf("error closing Redis: %v", err)
		} else {
			log.Println("redis connection closed")
		}
	}

}

func initPostgres(ctx context.Context) *gorm.DB {

	cfg := datastore.Config{
		Host:            viper.GetString("postgres.host"),
		Port:            viper.GetInt("postgres.port"),
		User:            viper.GetString("postgres.user"),
		Password:        viper.GetString("postgres.password"),
		DBName:          viper.GetString("postgres.dbname"),
		MaxIdleConns:    viper.GetInt("postgres.max_idle_conns"),
		MaxOpenConns:    viper.GetInt("postgres.max_open_conns"),
		ConnMaxLifetime: viper.GetDuration("postgres.conn_max_lifetime"),
		ConnMaxIdleTime: viper.GetDuration("postgres.conn_max_idle_time"),
		Debug:           viper.GetBool("postgres.debug"),
	}

	factory, err := datastore.NewDatastoreFactory(datastore.Postgres, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := factory.Connect(ctx); err != nil {
		log.Fatalf("Postgres connect fail: %v", err)
	}

	return factory.GetClient().(*gorm.DB)
}

func initRedis(ctx context.Context) *redis.Client {
	cfg := datastore.Config{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DBName:   viper.GetString("redis.dbname"),
	}

	factory, err := datastore.NewDatastoreFactory(datastore.Redis, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := factory.Connect(ctx); err != nil {
		log.Fatalf("Redis connect fail: %v", err)
	}

	return factory.GetClient().(*redis.Client)
}
