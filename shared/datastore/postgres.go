package datastore

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	cfg Config
	db  *gorm.DB
}

func (p *PostgresDB) Connect(ctx context.Context) error {

	logLevel := logger.Error

	if p.cfg.Debug {
		logLevel = logger.Info
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		p.cfg.Host, p.cfg.Port, p.cfg.User, p.cfg.Password, p.cfg.DBName)

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(p.cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(p.cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(p.cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(p.cfg.ConnMaxIdleTime)

	p.db = db

	return p.Ping(ctx)
}

func (p *PostgresDB) Ping(ctx context.Context) error {

	if p.db == nil {
		return fmt.Errorf("database not connected")
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (p *PostgresDB) Close(ctx context.Context) error {

	if p.db == nil {
		return nil
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (p *PostgresDB) GetClient() interface{} {
	return p.db
}
