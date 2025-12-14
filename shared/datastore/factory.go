package datastore

import (
	"context"
	"fmt"
	"time"
)

type DatabaseType string

const (
	Postgres DatabaseType = "postgres"
	Redis    DatabaseType = "redis"
)

type Datastore interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	GetClient() interface{}
}

type Config struct {
	DSN             string
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	Debug           bool
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewDatastoreFactory(dbType DatabaseType, cfg Config) (Datastore, error) {

	switch dbType {
	case Postgres:
		return &PostgresDB{cfg: cfg}, nil
	case Redis:
		return &RedisDB{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
