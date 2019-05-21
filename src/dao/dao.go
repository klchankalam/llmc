package dao

import (
	"entity"
	"fmt"
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func InitDB() {
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

type DAO interface {
	FindWithLimitAndOffset(db *gorm.DB, limit int, offset int, out *[]entity.Order)
	FindFirstWithIdAndStatus(db *gorm.DB, status string, id int, out *entity.Order)
	UpdateOrderStatus(db *gorm.DB, modelToUpdate *entity.Order, newStatus string, oldStatus string) *gorm.DB
	CreateOrder(db *gorm.DB, modelToCreate *entity.Order) *gorm.DB
}

type GormDB struct {
	DAO
}

func (gdb *GormDB) FindWithLimitAndOffset(db *gorm.DB, limit int, offset int, out *[]entity.Order) {
	db.Limit(limit).Offset(offset).Find(out)
}

func (gdb *GormDB) FindFirstWithIdAndStatus(db *gorm.DB, status string, id int, out *entity.Order) {
	db.Where("status = ?", status).First(out, id)
}

func (gdb *GormDB) UpdateOrderStatus(db *gorm.DB, modelToUpdate *entity.Order, newStatus string, oldStatus string) *gorm.DB {
	return db.Model(modelToUpdate).Where("Status = ?", oldStatus).Update("Status", newStatus)
}

func (gdb *GormDB) CreateOrder(db *gorm.DB, modelToCreate *entity.Order) *gorm.DB {
	return db.Create(modelToCreate)
}
