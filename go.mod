module kfc_promo_bot

go 1.17

require (
	github.com/glebarez/sqlite v1.6.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	gorm.io/gorm v1.24.2
	kfc_promo_bot/internal/pkg/kfcpromobotdatabase v0.0.1-unpublished
	kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain v0.0.1-unpublished
	kfc_promo_bot/internal/pkg/mediarepository v0.0.1-unpublished
	kfc_promo_bot/pkg/kfcapi v0.0.1-unpublished
	kfc_promo_bot/pkg/stringset v0.0.0-00010101000000-000000000000
)

require (
	github.com/glebarez/go-sqlite v1.20.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	kfc_promo_bot/internal/pkg/mediarepository/database v0.0.1-unpublished // indirect
	kfc_promo_bot/internal/pkg/mediarepository/database/domain v0.0.1-unpublished // indirect
	modernc.org/libc v1.21.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.4.0 // indirect
	modernc.org/sqlite v1.20.0 // indirect
)

replace (
	kfc_promo_bot/internal/pkg/kfcpromobotdatabase => ./internal/pkg/kfcpromobotdatabase
	kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain => ./internal/pkg/kfcpromobotdatabase/domain
	kfc_promo_bot/internal/pkg/mediarepository => ./internal/pkg/mediarepository
	kfc_promo_bot/internal/pkg/mediarepository/database => ./internal/pkg/mediarepository/database
	kfc_promo_bot/internal/pkg/mediarepository/database/domain => ./internal/pkg/mediarepository/database/domain
	kfc_promo_bot/pkg/kfcapi => ./pkg/kfcapi
	kfc_promo_bot/pkg/stringset => ./pkg/stringset
)
