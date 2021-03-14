package main

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	"github.com/Stepan1328/game-test-bot/game_logic"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func ActionsWithUpdates(updates *tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
	for update := range *updates {
		CheckUpdate(bot, &update)
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
	runGame := cust.Players[Message.From.ID].RunGame
	fmt.Println(Message.Text)

	if Message.Command() != "" && !runGame {
		RecognitionCommand(bot, update)
	}

	if Message.Command() == "stopgame" {
		fmt.Println("stopgame")
		StopGame(bot, Message)
		return
	}

	if runGame {
		UnfinishedGameMsg(bot, Message)
		return
	}
}

func CheckPlayer(bot *tgbotapi.BotAPI, update *tgbotapi.Update) bool {
	if update.Message != nil {
		if _, inBase := cust.Players[update.Message.From.ID]; !inBase {
			addToBase(update.Message.From.ID, update.Message.Chat.ID)
		}
	}

	if update.CallbackQuery != nil && cust.Players[update.CallbackQuery.From.ID].RunGame {
		cust.TranslateUpdate <- *update.CallbackQuery
		return true
	}

	if update.CallbackQuery != nil {
		ChangeLanguage(bot, update.CallbackQuery)
		return true
	}

	return false
}

func addToBase(PlayerID int, chatID int64) {
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

	game_logic.ParseLangMap(PlayerID)
}

func StopGame(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	if cust.Players[Message.From.ID].RunGame {
		cust.StopChannel <- *Message
	} else {
		msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.ID].Location.Dictionary["norungame"])

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func SetLanguage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, cust.Players[update.Message.From.ID].Location.Dictionary["change_lang"])

	ru := tgbotapi.NewInlineKeyboardButtonData("Русский", "ru")
	en := tgbotapi.NewInlineKeyboardButtonData("English", "en")
	row := tgbotapi.NewInlineKeyboardRow(ru, en)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	cust.OccupiedSells = append(cust.OccupiedSells, msgData.MessageID)
}

func ChangeLanguage(bot *tgbotapi.BotAPI, updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.ID
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

	go game_logic.DeleteMessage(bot, cust.Players[playerID].ChatID)
}

func ChangeSide(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	playerID := update.Message.From.ID

	if cust.Players[playerID].FirstMove {
		cust.Players[playerID].FirstMove = false
		msg.Text = "Now you play for ⭕"
	} else {
		cust.Players[playerID].FirstMove = true
		msg.Text = "Now you play for ❌"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func GameLaunch(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	playerID := update.Message.From.ID
	game_logic.Tttgame(bot, update)
	cust.Players[playerID].ChatID = update.Message.Chat.ID
	go game_logic.ListenCallbackQuery(update, bot)

	if _, ok := cust.Players[playerID]; ok {
		cust.Players[playerID].RunGame = true
	} else {
		log.Println("Failed to find user")
	}
}

func StartMsg(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.ID].Location.Dictionary["start"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func UnfinishedGameMsg(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.ID].Location.Dictionary["unfinished"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func RecognitionCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "setlanguage":
		SetLanguage(bot, update)
	case "changeside":
		ChangeSide(bot, update)
	case "tttgame":
		GameLaunch(bot, update)
	case "start":
		StartMsg(bot, update.Message)
	}
}
