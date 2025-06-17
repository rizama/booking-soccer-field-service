package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() (*gorm.DB, error) {
	config := Config

	encodedPassword := url.QueryEscape(config.Database.Password)
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Database.Username,
		encodedPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Failed to connect to database")
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatal("Failed to get database instance")
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConnection)
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConnection)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Database.MaxLifetimeConnection) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.Database.MaxIdleTime) * time.Second)

	return db, nil

}
