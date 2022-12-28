package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"kfc_promo_bot/internal/pkg/kfcpromobotdatabase"
	"kfc_promo_bot/internal/pkg/kfcpromobotdatabase/domain"
	"kfc_promo_bot/internal/pkg/mediarepository"
	"kfc_promo_bot/pkg/kfcapi"
	"kfc_promo_bot/pkg/stringset"

	// "gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var logger *log.Logger

var database kfcpromobotdatabase.KFCPromoBotDatabase
var botClient *tgbotapi.BotAPI
var apiClinet *kfcapi.Client
var mediaStorage *mediarepository.MediaRepository

const DEFAULT_DATABASE_PATH = "." + string(os.PathSeparator) + "database.db"
const DEFAULT_LOCAL_CACHE_PATH = "." + string(os.PathSeparator) + "cache"
const UPDATE_LIMIT = 55
const AUTOUPDATE_TIMEOUT_SEC = 600
const REJECT_COUPON_RU = "Данный купон не действителен"

func PromoInfoToMessage(promo kfcapi.Promo) ImageHashCaptionPair {
	var str strings.Builder

	value1 := promo.Value

	str.WriteString(value1.Translation.Ru.Title)
	str.WriteRune('\n')

	{
		price := value1.Price
		var discount float64
		discount = 100.0 - float64(value1.Price.Amount)/float64(value1.OldPrice)*100
		str.WriteString(fmt.Sprintf("*%d* ~%d~ %s (*%.2f%%*)\n", price.Amount / 100, value1.OldPrice / 100, price.Currency, discount))
	}

	for _, modifierGroup := range value1.ModifierGroups {
		str.WriteString(" - ")

		if strings.Contains(modifierGroup.Title.Ru, "Кофе") ||
		   strings.Contains(modifierGroup.Title.Ru, "кофе") ||
		   strings.Contains(modifierGroup.Title.Ru, "Соус") ||
		   strings.Contains(modifierGroup.Title.Ru, "соус") ||
		   strings.Contains(modifierGroup.Title.Ru, "Лимонад") {
			str.WriteString(modifierGroup.Title.Ru)
			if modifierGroup.UpLimit > 1 {
				str.WriteString(fmt.Sprintf(" (%d шт.)", modifierGroup.UpLimit))
			}
			str.WriteRune('\n')
			continue
		}

		if (len(modifierGroup.Modifiers) > 1) && !strings.Contains(modifierGroup.Title.Ru, "Блюдо") && false {
			str.WriteString(fmt.Sprintf("%s:\n", modifierGroup.Title.Ru))
		}

		modifiersCount := len(modifierGroup.Modifiers)
		for j, modifier := range modifierGroup.Modifiers {

			// modifierNameShort := strings.ReplaceAll(modifier.Title.Ru, (modifierGroup.Title.Ru + " ") ,"")

			modifierNameShort := modifier.Title.Ru

			modifierNameShort = strings.ReplaceAll(modifierNameShort, " Оригинальный", "")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, " оригинальный", "")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, " Оригинальных", "")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, " оригинальные", "")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, "Острый", "🌶")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, "острый", "🌶")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, "Острых", "🌶")
			modifierNameShort = strings.ReplaceAll(modifierNameShort, "острые", "🌶")

			str.WriteString(modifierNameShort)

			// required 2 of same type
			if len(modifierGroup.Modifiers) <= 1 && modifierGroup.UpLimit > 1 {
				str.WriteString((fmt.Sprintf(" (%d шт.)", modifierGroup.UpLimit)))
			}

			if j + 1 < modifiersCount {
				str.WriteString(" или ")
			}
		}
		str.WriteRune('\n')
	}

	res := str.String()

	res = strings.ReplaceAll(res, "(", "\\(")
	res = strings.ReplaceAll(res, ")", "\\)")
	res = strings.ReplaceAll(res, ".", "\\.")
	res = strings.ReplaceAll(res, "-", "\\-")
	res = strings.ReplaceAll(res, "+", "\\+")

	var promoImage string

	if strings.Contains(promo.Value.Translation.Ru.Title, "5050") {
		// this causes unique constrait fail
		promoImage = promo.Value.ModifierGroups[0].Modifiers[0].Media.Image
	} else {
		promoImage = promo.Value.Media.Image
	}

	return ImageHashCaptionPair{
		KfcImageHash: promoImage,
		Caption: res,
	}
}

// Compares db value and api value
// Considering both are valid structure jsons with common field order:
// i.e. : value, status, elapsed, createdAt
func CompareCodeInfoValue(codeInfo1 []byte, codeInfo2 []byte) int {
	s1 := string(codeInfo1)
	s2 := string(codeInfo2)

	s1 = s1[:strings.LastIndex(s1, string(`,"status"`))]
	s2 = s2[:strings.LastIndex(s2, string(`,"status"`))]

	return strings.Compare(s1, s2)
}

// Retrives promo info, updates database if required
// Returns nil if no such promo
func GetRestPromoInfoRawWithDBUpdate(restID string, code string) []byte {
	domainPromo := database.GetPromoByCodeOrNew(restID, code)

	checked := time.Now().Unix()

	if domainPromo.Checked >= (checked - UPDATE_LIMIT) {
		// last check was not too long ago
		// or
		// domainPromo just inited
		return domainPromo.Data
	}

	// time to check is data up to date
	codeInfoRaw, err := apiClinet.GetRestPromoInfoRaw(restID ,code)


	if err != nil {
		if err == kfcapi.ErrNotFound {
			// delete outdated entry if any
			database.DeletePromoIfAny(restID, code)
		} else if err == kfcapi.ErrExceedTryCountLimit {
			// Nothing
		} else {
			logger.Println("138:", err.Error())
		}
		return nil
	}

	if len(domainPromo.Data) > 0 && CompareCodeInfoValue(codeInfoRaw, domainPromo.Data) == 0 {
		// no need to update data db
		database.SetCheckedOnPromo(domainPromo, checked)
	} else {
		database.SetNewDataToPromo(domainPromo, codeInfoRaw, checked)
	}

	return codeInfoRaw
}

type ImageHashCaptionPair struct {
	KfcImageHash string
	Caption string
}

func PrepareAndSendPhotoMessage(kfcImageHash string, caption string, chatID int64) error {
	photoConfig, isImageNew, err  := mediaStorage.NewPhotoMsg(chatID, kfcImageHash)
	if err != nil {
		return err
	}

	photoConfig.Caption = caption
	photoConfig.ParseMode = "MarkdownV2"

	ret, err := botClient.Send(photoConfig)
	if err != nil {
		return err
	}

	// TODO log
	logger.Printf("Sent message in chat %d: \"%s\"\n", chatID, strings.ReplaceAll(ret.Caption, "\n", "\\n"))

	if isImageNew && len(*ret.Photo) > 0 {
		photoSize := *ret.Photo
		mediaStorage.AddCachedImage(kfcImageHash, photoSize[0].FileID)
	}

	return nil
}

func sendPhotoMessagesToChat(imageHashCaptionPairs []ImageHashCaptionPair, chatID int64) {
	for numberSent, imageHashCaptionPair := range imageHashCaptionPairs {
		if numberSent != 0 {
			time.Sleep(100 * time.Millisecond)
		}

		err := PrepareAndSendPhotoMessage(imageHashCaptionPair.KfcImageHash, imageHashCaptionPair.Caption, chatID)
		if err != nil {
			logger.Println(err.Error())
			continue
		}
		// TODO log
		// logger.Printf("Notified user [%d]\n", chatID)
	}
}

func DistributePhotoMessagesAmongChats(imageHashCaptionPairs []ImageHashCaptionPair, chatIDs[]int64) {
	for _, chatID := range chatIDs {
		go sendPhotoMessagesToChat(imageHashCaptionPairs, chatID)
	}
}

func UpdateKnownCodes() {
	logger.Println("Updating known codes...")

	restIDs := database.GetSubscribedUsersFavRests()

	for _, restID := range restIDs {

		res, err := apiClinet.GetRestPromoCodes(restID)
		if err != nil {
			// TODO log
			logger.Println(err.Error())
		}

		promoCodesToUpdate := stringset.New(res...)

		knownPromos := database.GetKnownRestPromos(restID)

		currentTime := time.Now().Unix()

		for _, promo := range knownPromos {
			if promo.Checked < (currentTime - UPDATE_LIMIT) {
				promoCodesToUpdate.Add(promo.Code)
			} else {
				promoCodesToUpdate.PureRemove(promo.Code)
			}
		}

		for _, code := range promoCodesToUpdate.ToSlice() {
			GetRestPromoInfoRawWithDBUpdate(restID, code)
		}
	}
}

func NotifyUsersIfNewPromos() {
	restIDs := database.GetSubscribedUsersFavRests()

	for _, restID := range restIDs {
		newPromos := database.GetNewRestPromos(restID)
		if len(newPromos) == 0 {
			return
		}

		var notificationMsgs []ImageHashCaptionPair
		for _, newPromo := range newPromos {
			var promo kfcapi.Promo
			json.Unmarshal(newPromo.Data , &promo)

			notificationMsg := PromoInfoToMessage(promo)
			notificationMsg.Caption = "🆕 " + notificationMsg.Caption

			notificationMsgs = append(notificationMsgs, notificationMsg)
		}

		database.SetPromosToOld(newPromos)

		users := database.GetSubscribedToRestUsers(restID)

		var chatIDs []int64
		for _, user := range users {
			chatIDs = append(chatIDs, user.ChatID)
		}

		DistributePhotoMessagesAmongChats(notificationMsgs, chatIDs)
	}
}

func UpdateAll() {

	UpdateKnownCodes()

	// At this point table up to date with website or even better
	// Now we can notify if where is any new codes
	NotifyUsersIfNewPromos()
}

func StartHandler(chatId int64, params []string) []tgbotapi.Chattable  {
	database.CreateNewUserIfNotExist(chatId)
	return nil
}

func StopHandler(chatId int64, params []string) []tgbotapi.Chattable  {
	database.DeleteUserIfExist(chatId)
	return nil
}

func SetFavRestHander(chatId int64, params []string) []tgbotapi.Chattable  {
	chat, ok := database.FindUserByChatID(chatId)
	if !ok {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chat.ChatID, "You have to type /start before using this command")}
	}

	favRestStatusMessage := func (chat domain.User) tgbotapi.MessageConfig {
		return tgbotapi.NewMessage(chat.ChatID, "Favourite Rest ID: " + chat.FavRestID )
	}

	if len(params) == 0 {
		return []tgbotapi.Chattable{favRestStatusMessage(chat)}
	}

	newFavRestID := params[0]

	if newFavRestID == chat.FavRestID {
		return []tgbotapi.Chattable{favRestStatusMessage(chat)}
	}

	database.SetUserFavRest(chat, newFavRestID)

	chat, ok = database.FindUserByChatID(chatId)
	if !ok {
		logger.Println("Somehow user was deleted after setting his subsription")
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chat.ChatID, "You have to type /start before using this command")}
	}

	return []tgbotapi.Chattable{favRestStatusMessage(chat)}
}

func NorificationsHandler(chatId int64, params []string) []tgbotapi.Chattable   {
	chat, ok := database.FindUserByChatID(chatId)
	if !ok {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chat.ChatID, "You have to type /start before using this command")}
	}

	newNotificationsStatusMessage := func (chat domain.User) tgbotapi.MessageConfig {
		trueFalseToOnOff := func(isTrue bool) string {
			if chat.Subscribed {
				return "on"
			}
			return "off"
		}

		return tgbotapi.NewMessage(chat.ChatID, "Notifications: " + trueFalseToOnOff(chat.Subscribed) )
	}

	if len(params) == 0 {
		return []tgbotapi.Chattable{newNotificationsStatusMessage(chat)}
	}

	var new_subscribed bool

	switch params[0] {
	case "on", "true":
		new_subscribed = true
		break

	case "off", "false":
		new_subscribed = false
		break

	case "toggle":
		new_subscribed = !chat.Subscribed
		break

	default:
		return []tgbotapi.Chattable{newNotificationsStatusMessage(chat)}
	}

	if new_subscribed == chat.Subscribed {
		return []tgbotapi.Chattable{newNotificationsStatusMessage(chat)}
	}

	database.SetUserSubscribed(chat, new_subscribed)
	chat, ok = database.FindUserByChatID(chatId)
	if !ok {
		logger.Println("Somehow user was deleted after setting his subsription")
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chat.ChatID, "You have to type /start before using this command")}
	}

	return []tgbotapi.Chattable{newNotificationsStatusMessage(chat)}
}

func PromoHandler(chatID int64, params []string) []tgbotapi.Chattable {

	if len(params) != 1 && len(params) != 2 {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, "usage: /promo <code> [restID]")}
	}

	var restID string
	var code string = params[0]

	if len(params) == 2 {
		restID = params[1]
	} else {
		u, found := database.FindUserByChatID(chatID)
		if !found {
			return []tgbotapi.Chattable{
				tgbotapi.NewMessage(chatID, "Not provided KFC's restourant ID you have to use /start to set favourite restourant"),
			}
		}
		if len(u.FavRestID) == 0 {
			return []tgbotapi.Chattable{
				tgbotapi.NewMessage(chatID, "Favourite restourant is not set use /setfavrest to set it"),
			}
		}

		restID = u.FavRestID
	}


	promoInfoRaw := GetRestPromoInfoRawWithDBUpdate(restID, code)

	if promoInfoRaw == nil {
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, REJECT_COUPON_RU)}
	}

	var promoInfo kfcapi.Promo
	json.Unmarshal(promoInfoRaw, &promoInfo)

	msg := PromoInfoToMessage(promoInfo)

	err := PrepareAndSendPhotoMessage(msg.KfcImageHash, msg.Caption, chatID)
	if err == nil {
		return nil
	}

	// TODO log
	logger.Println(err)
	return []tgbotapi.Chattable{tgbotapi.NewMessage(chatID, msg.Caption)}
}

func handleMessage(reqMsg *tgbotapi.Message, botClient *tgbotapi.BotAPI) {

	logger.Printf("Received a text message in chat %d [%s]:%s\n", reqMsg.Chat.ID, reqMsg.From.UserName, reqMsg.Text)

	args := strings.Split(reqMsg.Text, " ")

	prefix := "/"
	command := args[0]
	if (!strings.HasPrefix(command, prefix)) {
		return
	}
	command = strings.TrimPrefix(command, prefix)

	params := args[1:]

	var msgs []tgbotapi.Chattable

	var commandHandler func(chatID int64, params []string) []tgbotapi.Chattable

	switch command {

	case "promo":
		commandHandler = PromoHandler
		break

	case "start":
		commandHandler = StartHandler
		break

	case "stop":
		commandHandler = StopHandler
		break

	case "setfavrest":
		commandHandler = SetFavRestHander
		break

	case "notifications":
		commandHandler = NorificationsHandler
		break

	default:
		return
	}

	msgs = commandHandler(reqMsg.Chat.ID, params)

	for n, m := range msgs {
		if n != 0 {
			time.Sleep(1 * time.Second)
		}

		ret, _ := botClient.Send(m)

		// TODO log
		logger.Printf("Sent message in chat %d: \"%s\"\n", reqMsg.Chat.ID, strings.ReplaceAll(ret.Text, "\n", "\\n"))
	}
}

func Listener(botClient *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := botClient.GetUpdatesChan(u)

	for update := range updates {
		// Ignore any non-Message Updates
		if update.Message == nil {
			continue
		}

		go handleMessage(update.Message, botClient)
	}
}

func Updater() {
	var nextUpdateAt time.Time = time.Now()
	var sleepDuration time.Duration

	for {
		nextUpdateAt = nextUpdateAt.Add(time.Second * AUTOUPDATE_TIMEOUT_SEC)

		UpdateAll()

		sleepDuration = time.Until(nextUpdateAt)

		logger.Printf("Next update at: %v\n", nextUpdateAt.Format("2006/01/02 15:04:05"))
		time.Sleep(sleepDuration)
	}
}

func main() {

	logger = log.New(log.Default().Writer(), "[  General] ", log.LstdFlags | log.Lmsgprefix)


	dbPath := os.Getenv("BOTDATABASEPATH")
	if len(dbPath) == 0 {
		dbPath = DEFAULT_DATABASE_PATH
	}

	// Try open provided dbPath
	f, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logger.Printf("ERROR: can't open database: %s\n", err.Error())
		os.Exit(1)
	}
	f.Close()

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Println("ERROR: failed to connect to database:", err)
		os.Exit(1)
	}

	logger.Printf("using %s as database\n", dbPath)

	apiClinet, err = kfcapi.NewClientFromEnv(logger)
	if err != nil {
		logger.Println("ERROR: failed to initialize KFC API client:", err)
		os.Exit(1)
	}

	botClient, err = tgbotapi.NewBotAPI(os.Getenv("TGBOTTOKEN"))
	if err != nil {
		logger.Println("Failed to initialize bot API:", err)
		os.Exit(1)
	}

	database = kfcpromobotdatabase.New(logger, db)

	cachePath := os.Getenv("CACHEPATH")
	if len(cachePath) == 0 {
		cachePath = DEFAULT_LOCAL_CACHE_PATH
	}

	mediaLink := os.Getenv("MEDIALINK")
	if len(mediaLink) == 0 {
		logger.Println("ERROR: environment variable MEDIALINK is required")
		os.Exit(1)
	}

	mediaStorage = mediarepository.New(logger, db, cachePath, mediaLink)

	logger.Printf("Started at @%s\n", botClient.Self.UserName)

	go Updater()
	Listener(botClient)
}
