module mediarepository

go 1.17

require (
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	gorm.io/gorm v1.24.2
	kfc_promo_bot/internal/pkg/mediarepository/database v0.0.1-unpublished
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	kfc_promo_bot/internal/pkg/mediarepository/database/domain v0.0.1-unpublished // indirect
)

replace (
	kfc_promo_bot/internal/pkg/mediarepository/database v0.0.1-unpublished => ./database
	kfc_promo_bot/internal/pkg/mediarepository/database/domain v0.0.1-unpublished => ./database/domain
)
