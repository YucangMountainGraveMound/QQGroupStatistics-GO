package db

import (
	"fmt"

	"dormon.net/qq/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*gorm.DB
}

var db *gorm.DB

func InitialDB() {
	var err error
	db, err = connectDB()
	if err != nil {
		logrus.Fatal(err)
	}
}

func connectDB() (*gorm.DB, error) {

	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
			config.Config().DatabaseConfig.Host,
			config.Config().DatabaseConfig.Username,
			config.Config().DatabaseConfig.DatabaseName,
			config.Config().DatabaseConfig.Password,
		))

	if err != nil {
		return nil, err
	}
	return db, nil
}

func CloseDB() {
	db.Close()
}

func GetDB() *gorm.DB {
	return db
}
