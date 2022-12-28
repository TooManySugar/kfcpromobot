package database

import (
	"gorm.io/gorm"
	"log"
	domain "kfc_promo_bot/internal/pkg/mediarepository/database/domain"
)

type DB struct {
	logger *log.Logger
	real_db *gorm.DB
}

func NewDB(logger *log.Logger, database *gorm.DB) DB {
	database.AutoMigrate(&domain.CachedImage{})
	var res DB
	res = DB {
		logger: log.New(logger.Writer(), "[  MediaDB] ", log.LstdFlags | log.Lmsgprefix),
		real_db: database,
	}
	return res
}

func (this DB) GetCachedTgImageID(kfcHash string) (tgPhotoID string, ok bool) {
	cachedImage := domain.CachedImage {
		KfcHash: kfcHash,
	}
	if this.real_db.Where(&cachedImage).Find(&cachedImage).RowsAffected == 0 {
		return "", false
	}

	return cachedImage.TgPhotoID, true
}

func (this DB) AddCachedImage(kfcHash string, tgPhotoID string) (ok bool) {
	cachedImage := domain.CachedImage {
		KfcHash: kfcHash,
		TgPhotoID: tgPhotoID,
	}

	if err := this.real_db.Save(&cachedImage).Error; err != nil {
		this.logger.Printf("ERROR: can't save new cached image: %s\n", err.Error())
		return false
	}

	return true
}