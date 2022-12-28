
# KFC Promo Bot

This is telegram bot to query Russian's KFC api and notify if where is a new coupons available

## Disclaimer 

It's mostly me doing thing for myself and for practice so don't judge or blame me if you try use it and something not working

## Quick Start

Compile it for you desired target platform

Set following environment variables (it's required ones):

```
TGBOTTOKEN - Telegram token from @BotFather
KFCAPILINK - URL to KFC's api
MEDIALINK  - URL to KFC's media api
```

In it's working directory bot will create `cache` folder and `database.db` - SQLite database

## Docker

To build docker image root certificates required. You can obtain them using [this](https://github.com/agl/extract-nss-root-certs). Those must be called `ca_certs.pem` and poot in same folder with `Dockerfile` or edit `Dockerfile` if you so smart.

```bash
docker build -t desiredtagforimage .
```

Before running container create volume to store cache, database, etc. It must be mounted at `/app` (could be changed in Dockerfile)

docker run example:

```bash
docker run -d --name kfc_promo_bot --mount source=bot_volume,target=/app \
    -e "TGBOTTOKEN=1234567890:Loremipsumdolorsitametconsecteturad" \
    -e "KFCAPILINK=https://..." \
    -e "MEDIALINK=https://..." \
    desiredtagforimage
```

## Other environment variables

```
BOTDATABASEPATH - path to database default: ./database.db
CACHEPATH       - path to cache directory default: ./cache
```

## Bot commands

`/start` - Start bot

`/stop` - Stop bot (drops all chat settings)

`/promo <code> [<rest_id>]` - get info on `code`. If `rest_id` provided returns info on code for specific restoraunt

`/setfavrest [<rest_id>]` - setting id of favorite restoraunt (id could be found somethere on kfc.ru website). If sent returns last set value

`/notifications [on/off/true/false/toggle]` - if sent without arguments returns current state of notifications in current channel\
NOTE: `/setfavrest` must be set for notifactions to work

