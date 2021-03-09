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
		irrelevantField(bot, update.CallbackQuery.Message.Chat.ID)
		return
	}
	runGame := cust.Players[Message.From.ID].RunGame
	fmt.Println(Message.Text)

	if Message.Command() != "" && !runGame {
		RecognitionCommand(update, bot)
	}

	if Message.Command() == "stopgame" {
		fmt.Println("stopgame")
		StopGame(Message, bot)
		return
	}

	if runGame {
		UnfinishedGameMsg(Message.Chat.ID, bot)
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

	return false
}

func addToBase(PlayerID int) {
	cust.Players[PlayerID] = &cust.UsersStatistic{
		Field: &cust.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}
}

func StopGame(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	if cust.Players[Message.From.ID].RunGame {
		cust.StopChannel <- *Message
	} else {
		msg := tgbotapi.NewMessage(Message.Chat.ID, cust.LangMap["norungame"])

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func ChangeSide(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	if cust.Players[update.Message.From.ID].FirstMove {
		cust.Players[update.Message.From.ID].FirstMove = false
		msg.Text = "Now you play for ⭕"
	} else {
		cust.Players[update.Message.From.ID].FirstMove = true
		msg.Text = "Now you play for ❌"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func GameLaunch(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	msgID := game_logic.Tttgame(update, bot)
	cust.Players[update.Message.From.ID].ChatID = update.Message.Chat.ID
	go game_logic.ListenCallbackQuery(update, bot, msgID)

	if _, ok := cust.Players[update.Message.From.ID]; ok {
		cust.Players[update.Message.From.ID].RunGame = true
	} else {
		log.Println("Failed to find user")
	}
}

func StartMsg(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.LangMap["start"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func UnfinishedGameMsg(ChatID int64, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(ChatID, cust.LangMap["unfinished"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func irrelevantField(bot *tgbotapi.BotAPI, ChatID int64) {
	msg := tgbotapi.NewMessage(ChatID, cust.LangMap["irrelevant_field"])

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func RecognitionCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch update.Message.Command() {
	case "changeside":
		ChangeSide(update, bot)
	case "tttgame":
		GameLaunch(update, bot)
	case "start":
		StartMsg(update.Message, bot)
	}
}
