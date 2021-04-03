package main

import (
	"log"
	"strings"

	"github.com/Stepan1328/game-test-bot/clients"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Settings is a main function to open settings
func Settings(update *tgbotapi.Update) {
	playerID := update.Message.From.ID
	msg := tgbotapi.NewMessage(clients.Players[playerID].ChatID,
		"loading")

	loading := tgbotapi.NewInlineKeyboardButtonData("loading", " ")
	row := tgbotapi.NewInlineKeyboardRow(loading)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	msgData, err := clients.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	clients.Players[playerID].SettingID = msgData.MessageID
	editSettingMsg(update.Message.From.ID)
}

func editSettingMsg(playerID int) {
	dictionary := clients.Players[playerID].Location.Dictionary

	var side string
	if clients.Players[playerID].FirstMove {
		side = dictionary["settings_changeside_cross"]
	} else {
		side = dictionary["settings_changeside_zero"]
	}
	changeside := tgbotapi.NewInlineKeyboardButtonData(side, "settings/changeside")
	row1 := tgbotapi.NewInlineKeyboardRow(changeside)

	lang := tgbotapi.NewInlineKeyboardButtonData(dictionary["settings_set_lang"], "settings/language")
	var notification string
	if clients.Players[playerID].BattleInvite {
		notification = dictionary["settings_notification_on"]
	} else {
		notification = dictionary["settings_notification_off"]
	}
	invite := tgbotapi.NewInlineKeyboardButtonData(notification, "settings/notification")
	row2 := tgbotapi.NewInlineKeyboardRow(lang, invite)

	exit := tgbotapi.NewInlineKeyboardButtonData(dictionary["settings_exit"], "settings/exit")
	row3 := tgbotapi.NewInlineKeyboardRow(exit)

	markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3)
	editConfig := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      clients.Players[playerID].ChatID,
			MessageID:   clients.Players[playerID].SettingID,
			ReplyMarkup: &markup,
		},
		Text: clients.Players[playerID].Location.Dictionary["settings"],
	}

	_, err := clients.Bot.Send(editConfig)
	if err != nil {
		log.Println(err)
	}
}

// AnalyzeSettingRequest is a function which responds to the user's click
// on the settings fields
func AnalyzeSettingRequest(callbackQuery *tgbotapi.CallbackQuery) {
	data := strings.Replace(callbackQuery.Data, "settings/", "", 1)
	callbackQuery.Data = data
	switch strings.Split(data, "/")[0] {
	case "changeside":
		analyzeSideRequest(callbackQuery)
	case "language":
		analyzeLangRequest(callbackQuery)
	case "notification":
		analyzeNotificationRequest(callbackQuery)
	case "exit":
		DeleteSettingMsg(callbackQuery.From.ID)
	}
}

func analyzeSideRequest(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	if strings.Contains(callbackQuery.Data, "/") {
		callbackQuery.Data = strings.Replace(callbackQuery.Data, "changeside/", "", 1)
		analyzeSideAnswer(callbackQuery)
		return
	}

	cross := tgbotapi.NewInlineKeyboardButtonData("❌", "settings/changeside/cross")
	zero := tgbotapi.NewInlineKeyboardButtonData("⭕", "settings/changeside/zero")
	row1 := tgbotapi.NewInlineKeyboardRow(cross, zero)

	exit := tgbotapi.NewInlineKeyboardButtonData(clients.Players[playerID].Location.Dictionary["settings_back"], "settings/changeside/exit")
	row2 := tgbotapi.NewInlineKeyboardRow(exit)

	markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	editConfig := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      clients.Players[playerID].ChatID,
			MessageID:   clients.Players[playerID].SettingID,
			ReplyMarkup: &markup,
		},
		Text: clients.Players[playerID].Location.Dictionary["settings_changeside"],
	}

	_, err := clients.Bot.Send(editConfig)
	if err != nil {
		log.Println(err)
	}
}

func analyzeSideAnswer(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	switch callbackQuery.Data {
	case "cross":
		clients.Players[playerID].FirstMove = true
	case "zero":
		clients.Players[playerID].FirstMove = false
	}
	editSettingMsg(playerID)
}

func analyzeLangRequest(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	if strings.Contains(callbackQuery.Data, "/") {
		callbackQuery.Data = strings.Replace(callbackQuery.Data, "language/", "", 1)
		analyzeLangAnswer(callbackQuery)
		return
	}

	en := tgbotapi.NewInlineKeyboardButtonData("English 🇬🇧", "settings/language/en")
	ru := tgbotapi.NewInlineKeyboardButtonData("Русский 🇷🇺", "settings/language/ru")
	row1 := tgbotapi.NewInlineKeyboardRow(en, ru)

	exit := tgbotapi.NewInlineKeyboardButtonData(clients.Players[playerID].Location.Dictionary["settings_back"], "settings/language/exit")
	row2 := tgbotapi.NewInlineKeyboardRow(exit)

	markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	editConfig := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      clients.Players[playerID].ChatID,
			MessageID:   clients.Players[playerID].SettingID,
			ReplyMarkup: &markup,
		},
		Text: clients.Players[playerID].Location.Dictionary["settings_language"],
	}

	_, err := clients.Bot.Send(editConfig)
	if err != nil {
		log.Println(err)
	}
}

func analyzeLangAnswer(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	switch callbackQuery.Data {
	case "en":
		clients.Players[playerID].Location.Language = "en"
		clients.Players[playerID].ParseLangMap()
	case "ru":
		clients.Players[playerID].Location.Language = "ru"
		clients.Players[playerID].ParseLangMap()
	}
	editSettingMsg(playerID)
}

func analyzeNotificationRequest(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	if strings.Contains(callbackQuery.Data, "/") {
		callbackQuery.Data = strings.Replace(callbackQuery.Data, "notification/", "", 1)
		analyzeNotificationAnswer(callbackQuery)
		return
	}

	var data1, data2 string
	switch clients.Players[playerID].Location.Language {
	case "ru":
		data1 = "Включить уведомления 🔔"
		data2 = "Выключить уведомления 🔕"
	case "en":
		data1 = "Enable notifications 🔔"
		data2 = "Disable notifications 🔕"
	}
	on := tgbotapi.NewInlineKeyboardButtonData(data1, "settings/notification/on")
	off := tgbotapi.NewInlineKeyboardButtonData(data2, "settings/notification/off")
	row1 := tgbotapi.NewInlineKeyboardRow(on, off)

	exit := tgbotapi.NewInlineKeyboardButtonData(clients.Players[playerID].Location.Dictionary["settings_back"], "settings/notification/exit")
	row2 := tgbotapi.NewInlineKeyboardRow(exit)

	markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	editConfig := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      clients.Players[playerID].ChatID,
			MessageID:   clients.Players[playerID].SettingID,
			ReplyMarkup: &markup,
		},
		Text: clients.Players[playerID].Location.Dictionary["settings_notification"],
	}

	_, err := clients.Bot.Send(editConfig)
	if err != nil {
		log.Println(err)
	}
}

func analyzeNotificationAnswer(callbackQuery *tgbotapi.CallbackQuery) {
	playerID := callbackQuery.From.ID
	switch callbackQuery.Data {
	case "on":
		clients.Players[playerID].BattleInvite = true
	case "off":
		clients.Players[playerID].BattleInvite = false
	}
	editSettingMsg(playerID)
}

// DeleteSettingMsg is a function that deletes the settings message
func DeleteSettingMsg(playerID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(clients.Players[playerID].ChatID,
		clients.Players[playerID].SettingID)

	_, err := clients.Bot.Send(deleteMsg)
	if err != nil {
		log.Println(err)
	}
	clients.Players[playerID].SettingID = 0
}
