package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	DB *gorm.DB
}

var AppConfig Config

func InitializeDB() error {
	dsn := "root:hqy25257758@tcp(127.0.0.1:3306)/thebase?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	AppConfig.DB = db
	return nil
}
