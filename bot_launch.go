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
		CheckUpdate(&update, bot)
	}
}

func CheckUpdate(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if CheckPlayer(update) {
		return
	}

	Message := update.Message
	if Message == nil {
		irrelevantField(bot, update.CallbackQuery.Message)
		return
	}
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

func CheckPlayer(update *tgbotapi.Update) bool {
	if update.Message != nil {
		if _, inBase := cust.Players[update.Message.From.ID]; !inBase {
			addToBase(update.Message.From.ID)
		}
	}

	if update.CallbackQuery != nil && cust.Players[update.CallbackQuery.From.ID].RunGame {
		cust.TranslateUpdate <- *update.CallbackQuery
		return true
	}

	if update.CallbackQuery != nil {
		ChangeLanguage(update.CallbackQuery)
		return true
	}

	return false
}

func addToBase(PlayerID int) {
	cust.Players[PlayerID] = &cust.UsersStatistic{
		Location: &cust.Localization{
			Language: "en",
		},
		Field: &cust.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	game_logic.ParseMap(PlayerID)
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

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func ChangeLanguage(updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.ID
	switch updateCallback.Data {
	case "ru":
		cust.Players[playerID].Location.Language = "ru"
	case "en":
		cust.Players[playerID].Location.Language = "en"
	default:
	}
	game_logic.ParseMap(playerID)
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

	msgID := game_logic.Tttgame(bot, update)
	cust.Players[update.Message.From.ID].ChatID = update.Message.Chat.ID
	go game_logic.ListenCallbackQuery(update, bot, msgID)

	if _, ok := cust.Players[update.Message.From.ID]; ok {
		cust.Players[update.Message.From.ID].RunGame = true
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

func irrelevantField(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.ID].Location.Dictionary["irrelevant_field"])

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
