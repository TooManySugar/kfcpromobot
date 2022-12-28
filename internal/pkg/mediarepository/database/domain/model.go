package model

type CachedImage struct {
	ID uint `gorm:"primaryKey"`
	KfcHash string `gorm:"uniqueIndex"`
	TgPhotoID string `gorm:"uniqueIndex"`
}
