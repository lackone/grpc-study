package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB = InitDB()
)

func InitDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	//db.Logger = logger.Default.LogMode(logger.Info)
	return db
}
