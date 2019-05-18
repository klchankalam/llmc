package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func InitDb() {
	//db, err := gorm.Open("mysql", "user:password@tcp(db:3306)/db?charset=utf8mb4&parseTime=True")
	d, err := gorm.Open("postgres", "host=db port=5432 user=postgres dbname=postgres password=password sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %s", err.Error()))
	}
	db = d
}

func GetDB() *gorm.DB {
	return db
}
