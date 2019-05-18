package entity

import "time"

type Order struct {
	ID          uint64    `gorm:"primary_key" json:"id"`
	Distance    uint64    `gorm:"not null" json:"distance"`
	Status      string    `gorm:"type:varchar(10);not null" json:"status"`
	OriginsLat  string    `json:"-"`
	OriginsLong string    `json:"-"`
	DestLat     string    `json:"-"`
	DestLong    string    `json:"-"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}
