package domain

import (
	// ""
)

type Promo struct {
	ID uint `gorm:"primaryKey"`
	RestID string `gorm:"default:74321714"`
	Code string
	Data []byte
	New bool `gorm:"default:true"`
	BackUpdated int64
	Checked int64
}

type User struct {
	ID uint `gorm:"primaryKey"`
	ChatID int64 `gorm:"unique"`
	FavRestID string
	Subscribed bool
	CreatedAt int64
}
