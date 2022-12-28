module kfcpromobotdatabase

go 1.17

replace kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain v0.0.1-unpublished => ./domain

require (
    gorm.io/gorm v1.24.2
    kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain v0.0.1-unpublished
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
)
