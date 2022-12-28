package mediarepository

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"

	db "kfc_promo_bot/internal/pkg/mediarepository/database"
)

type MediaRepository struct {
	logger *log.Logger
	database db.DB
	localCachePath string
	mediaURL string
}

func New(logger *log.Logger, database *gorm.DB, localCachePath string, mediaURL string) *MediaRepository {

	thisLogger := log.New(logger.Writer(), "[MediaRepo] ", log.LstdFlags | log.Lmsgprefix)

	thisLogger.Printf("using %s as media cache storage\n", localCachePath)
	f, err := os.Stat(localCachePath)
	if err != nil {
		// try to create folder
		if os.Mkdir(localCachePath, 0777) != nil {
			thisLogger.Printf("ERROR: can't create %s for local cache", localCachePath)
			os.Exit(1)
		}
	} else if !f.IsDir() {
		thisLogger.Printf("ERROR: %s is not a directory", localCachePath)
		os.Exit(1)
	}

	return &MediaRepository{
		logger: thisLogger,
		database: db.NewDB(logger, database),
		localCachePath: localCachePath,
		mediaURL: mediaURL,
	}
}

func (this *MediaRepository) GetCachedTgImageID(kfcHash string) (tgPhotoID string, ok bool) {
	return this.database.GetCachedTgImageID(kfcHash)
}

func (this *MediaRepository) AddCachedImage(kfcHash string, tgPhotoID string) (ok bool) {
	return this.database.AddCachedImage(kfcHash, tgPhotoID)
}

func (this *MediaRepository) newTgPhotoUpload(chatID int64, imageData []byte) (photoConfig tgbotapi.PhotoConfig) {
	return tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: imageData,
	})
}

func (this *MediaRepository) NewPhotoMsg(chatID int64, kfcHash string) (photoConfig tgbotapi.PhotoConfig, isNew bool, err error) {

	// Search in cache
	if photoID, ok := this.database.GetCachedTgImageID(kfcHash); ok == true {
		return tgbotapi.NewPhotoShare(chatID, photoID), false, nil
	}

	// Search in local cache
	if imageData, err := ioutil.ReadFile(fmt.Sprintf("%s%c%s.webp", this.localCachePath, os.PathSeparator, kfcHash)); err == nil {
		return this.newTgPhotoUpload(chatID, imageData), true, nil
	}

	// Download from media server
	{
		var resp *http.Response
		if resp, err = http.Get(fmt.Sprintf("%s/%s", this.mediaURL, kfcHash)); err == nil {
			defer resp.Body.Close()
			var imageData []byte
			if imageData, err = io.ReadAll(resp.Body); err == nil {
				err = ioutil.WriteFile(fmt.Sprintf("%s%c%s.webp", this.localCachePath,  os.PathSeparator, kfcHash), imageData, 0777)
				if err != nil {
					this.logger.Println(err.Error())
				}
				return this.newTgPhotoUpload(chatID, imageData), true, nil
			}
		}
		this.logger.Printf("ERROR: can't get image \"%s\" from media server: %s\n", kfcHash, err.Error())
	}

	// Give up
	return tgbotapi.PhotoConfig{}, false, fmt.Errorf("can't make new photo msg")
}
