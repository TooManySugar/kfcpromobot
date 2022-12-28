package kfcpromobotdatabase

import (
	"log"
	"time"
	"gorm.io/gorm"
	"kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain"
)

type KFCPromoBotDatabase struct {
	logger *log.Logger
	db *gorm.DB
}

func New(logger *log.Logger, database *gorm.DB) KFCPromoBotDatabase {
	thisLogger := log.New(logger.Writer(), "[ Database] ", log.LstdFlags | log.Lmsgprefix)
	database.AutoMigrate(&domain.Promo{}, &domain.User{})
	return KFCPromoBotDatabase {
		logger: thisLogger,
		db: database,
	}
}

func (this KFCPromoBotDatabase) GetPromoByCodeOrNew(restID string, code string ) domain.Promo {
	var domainPromo domain.Promo
	domainPromo.Code = code
	domainPromo.RestID = restID
	this.db.FirstOrInit(&domainPromo, domainPromo)
	return domainPromo
}

func (this KFCPromoBotDatabase) DeletePromoIfAny(restID string, code string) {
	if this.db.Where(&domain.Promo{Code: code}).Delete(&domain.Promo{}).RowsAffected > 0 {
		this.logger.Println("deleted promo", code)
	}
}

func (this KFCPromoBotDatabase) SetCheckedOnPromo(domainPromo domain.Promo, checked int64) {
	// UPDATE `promos` SET `checked`={checked} WHERE `id` = {domainPromo.ID};
	domainPromo.Checked = checked
	this.db.Model(&domainPromo).Updates(&domain.Promo{Checked: checked})
}

func (this KFCPromoBotDatabase) SetNewDataToPromo(domainPromo domain.Promo, data []byte, checked int64) {
	this.logger.Println("New data arrived to code", domainPromo.Code)
	domainPromo.Data = data
	domainPromo.BackUpdated = checked
	domainPromo.Checked = checked
	domainPromo.New = true
	this.db.Save(&domainPromo)
}

func (this KFCPromoBotDatabase) GetKnownRestPromos(restID string) []domain.Promo {
	var knownPromos []domain.Promo
	this.db.Where(&domain.Promo{RestID: restID}).Find(&knownPromos)
	return knownPromos;
}

func (this KFCPromoBotDatabase) GetNewRestPromos(restID string) []domain.Promo {
	var newPromos []domain.Promo
	this.db.Where(&domain.Promo{New: true, RestID: restID}).Find(&newPromos)
	return newPromos
}

func (this KFCPromoBotDatabase) GetSubscribedUsersFavRests() []string {
	var favRests []string
	this.db.Model(&domain.User{}).Distinct("fav_rest_id").Where(&domain.User{Subscribed: true}).Find(&favRests)
	return favRests
}

func (this KFCPromoBotDatabase) GetSubscribedToRestUsers(restID string) []domain.User {
	var users []domain.User
	this.db.Where(&domain.User{Subscribed: true, FavRestID: restID}).Find(&users)
	return users
}

func (this KFCPromoBotDatabase) SetPromosToOld(promos []domain.Promo) {
	var promosIDs []uint
	for _, promo := range promos {
		promosIDs = append(promosIDs, promo.ID)
	}
	this.db.Model(domain.Promo{}).Where("id IN (?)", promosIDs).Update("new", 0)
}

func (this KFCPromoBotDatabase) CreateNewUserIfNotExist(chatID int64) {
	user := domain.User{ChatID: chatID}
	this.db.FirstOrInit(&user)
	if user.CreatedAt == 0 {
		user.CreatedAt = time.Now().Unix()
		this.db.Save(&user)
	}
}

func (this KFCPromoBotDatabase) DeleteUserIfExist(chatID int64) {
	user := domain.User{ChatID: chatID}
	this.db.FirstOrInit(&user)
	if user.CreatedAt != 0 {
		this.db.Delete(&user)
	}
}

func (this KFCPromoBotDatabase) FindUserByChatID(chatID int64) (user domain.User, found bool) {
	user = domain.User{ ChatID: chatID }
	if this.db.First(&user).Error == gorm.ErrRecordNotFound {
		return user, false
	}

	return user, true
}

func (this KFCPromoBotDatabase) SetUserSubscribed(user domain.User, subscirbed bool) {
	user.Subscribed = subscirbed
	this.db.Save(&user)
}

func (this KFCPromoBotDatabase) SetUserFavRest(user domain.User, favRestID string) {
	user.FavRestID = favRestID
	this.db.Save(&user)
}
