package main

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	"github.com/Stepan1328/game-test-bot/database"
	"github.com/Stepan1328/game-test-bot/game_logic"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

func ActionsWithUpdates(updates *tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
	for update := range *updates {
		go CheckUpdate(bot, &update)
	}
}

func CheckUpdate(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if CheckPlayer(bot, update) {
		return
	}

	Message := update.Message
	runGame := cust.Players[Message.From.UserName].RunGame
	fmt.Println(Message.Text)

	if Message.Command() != "" && !runGame {
		RecognitionCommand(bot, update)
	}

	if Message.Command() == "stopgame" {
		fmt.Println("stopgame")
		StopGame(bot, update)
		return
	}

	if runGame {
		UnfinishedGameMsg(bot, Message)
		return
	}
}

func CheckPlayer(bot *tgbotapi.BotAPI, update *tgbotapi.Update) bool {
	if update.Message != nil {
		if _, inBase := cust.Players[update.Message.From.UserName]; !inBase {
			addToBase(update.Message.From.UserName, update.Message.Chat.ID)
		}
	}

	if update.CallbackQuery != nil {
		if update.CallbackQuery.From.UserName == "" {
			NoneUserNamePlayer(bot, update.CallbackQuery.Message.Chat.ID)
			return true
		}

		if _, inBase := cust.Players[update.CallbackQuery.From.UserName]; !inBase {
			addToBase(update.CallbackQuery.From.UserName, update.CallbackQuery.Message.Chat.ID)
		}
	}

	if update.CallbackQuery != nil && cust.Players[update.CallbackQuery.From.UserName].RunGame {
		go game_logic.ListenCallbackQuery(bot, update)
		cust.TranslateUpdate <- *update.CallbackQuery
		return true
	}

	if update.CallbackQuery != nil {
		if _, inBase := cust.Battles[update.CallbackQuery.From.UserName]; inBase {
			if cust.Battles[update.CallbackQuery.From.UserName].RunGame {
				cust.TranslateBattle <- *update.CallbackQuery
				return true
			}
		}
	}

	if update.CallbackQuery != nil {
		ChangeLanguage(bot, update.CallbackQuery)
		return true
	}

	return false
}

func addToBase(PlayerID string, chatID int64) {
	cust.Players[PlayerID] = &cust.UsersStatistic{
		ChatID: chatID,
		Location: &cust.Localization{
			Language: "en",
		},
		Field: &cust.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	database.SaveBase()
	game_logic.ParseLangMap(PlayerID)
}

func NoneUserNamePlayer(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, but you don't have a Username in your telegram profile\nAdd it by going to Settings -> Edit -> Username")

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func StopGame(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if cust.Players[update.Message.From.UserName].RunGame {
		go game_logic.ListenCallbackQuery(bot, update)
		cust.StopChannel <- *update.Message
	} else {
		msg := tgbotapi.NewMessage(cust.Players[update.Message.From.UserName].ChatID, cust.Players[update.Message.From.UserName].Location.Dictionary["norungame"])

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func SetLanguage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, playerID string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, cust.Players[update.Message.From.UserName].Location.Dictionary["change_lang"])

	ru := tgbotapi.NewInlineKeyboardButtonData("Русский", "ru")
	en := tgbotapi.NewInlineKeyboardButtonData("English", "en")
	row := tgbotapi.NewInlineKeyboardRow(ru, en)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	cust.Players[playerID].OccupiedSells = append(cust.Players[playerID].OccupiedSells, msgData.MessageID)
}

func ChangeLanguage(bot *tgbotapi.BotAPI, updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.UserName
	switch updateCallback.Data {
	case "ru":
		cust.Players[playerID].Location.Language = "ru"
	case "en":
		cust.Players[playerID].Location.Language = "en"
	default:
		msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["finished_game"])

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}
	game_logic.ParseLangMap(playerID)
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["lang_set"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}

	database.SaveBase()
	go game_logic.DeleteMessage(bot, playerID, cust.Players[playerID].ChatID)
}

func ChangeSide(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	playerID := update.Message.From.UserName

	if cust.Players[playerID].FirstMove {
		cust.Players[playerID].FirstMove = false
		msg.Text = cust.Players[playerID].Location.Dictionary["play_zero"]
	} else {
		cust.Players[playerID].FirstMove = true
		msg.Text = cust.Players[playerID].Location.Dictionary["play_cross"]
	}

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
	database.SaveBase()
}

func GameLaunch(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	playerID := update.Message.From.UserName
	game_logic.Tttgame(bot, update)
	cust.Players[playerID].ChatID = update.Message.Chat.ID
	database.SaveBase()
	go game_logic.ListenCallbackQuery(bot, update)

	if _, ok := cust.Players[playerID]; ok {
		cust.Players[playerID].RunGame = true
	} else {
		log.Println("Failed to find user")
	}
	database.SaveBase()
}

func BattleLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)
	fmt.Println(update.Message.From.UserName)

	UserName := strings.Trim(update.Message.Text, "/tttbattle @")
	fmt.Println(UserName)
}

func StartMsg(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.UserName].Location.Dictionary["start"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func UnfinishedGameMsg(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.UserName].Location.Dictionary["unfinished"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func RecognitionCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "setlanguage":
		SetLanguage(bot, update, update.Message.From.UserName)
	case "changeside":
		ChangeSide(bot, update)
	case "tttgame":
		GameLaunch(bot, update)
	case "tttbattle":
		BattleLaunch(update)
	case "start":
		StartMsg(bot, update.Message)
	}
}
