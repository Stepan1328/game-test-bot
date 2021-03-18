package main

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	"github.com/Stepan1328/game-test-bot/game_logic"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"strings"
)

func ActionsWithUpdates(updates *tgbotapi.UpdatesChannel) {
	for update := range *updates {
		go CheckUpdate(&update)
	}
}

func CheckUpdate(update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if CheckPlayer(update) {
		return
	}

	Message := update.Message
	runGame := cust.Players[Message.From.UserName].RunGame
	fmt.Println(Message.Text)

	if Message.Command() == "stopgame" {
		StopGame(update)
		return
	}

	if Message.Command() != "" && !runGame {
		RecognitionCommand(update)
	}

	if runGame {
		UnfinishedGameMsg(Message)
		return
	}
}

func CheckPlayer(update *tgbotapi.Update) bool {
	// check username and is it in the base via Message
	if update.Message != nil {
		if update.Message.From.UserName == "" {
			NoneUserNamePlayer(update.Message.Chat.ID)
			fmt.Println("NoneUserNamePlayer")
			return true
		}

		if _, inBase := cust.Players[update.Message.From.UserName]; !inBase {
			addToPlayerBase(update.Message.From.UserName, update.Message.Chat.ID)
			fmt.Println("PLayer added to player base")
		}
	}

	// check username and is it in the base via Callback
	if update.CallbackQuery != nil {
		if update.CallbackQuery.From.UserName == "" {
			NoneUserNamePlayer(update.CallbackQuery.Message.Chat.ID)
			fmt.Println("NoneUserNamePlayer")
			return true
		}

		if _, inBase := cust.Players[update.CallbackQuery.From.UserName]; !inBase {
			addToPlayerBase(update.CallbackQuery.From.UserName, update.CallbackQuery.Message.Chat.ID)
			fmt.Println("PLayer added to player base")
		}
	}

	// translate gameUpdate (check run game and check Unlock mutex)
	if update.CallbackQuery != nil && cust.Players[update.CallbackQuery.From.UserName].RunGame {
		if !cust.Players[update.CallbackQuery.From.UserName].Field.Mutex {
			return true
		}
		go game_logic.ListenCallbackQuery(update)
		cust.TranslateUpdate <- *update.CallbackQuery
		return true
	}

	// translate battleUpdate (check run game and check Unlock mutex)
	if update.CallbackQuery != nil {
		if _, inBase := cust.Battles[update.CallbackQuery.From.UserName]; inBase {
			if cust.Players[update.CallbackQuery.From.UserName].RunGame { // TODO: mutex
				cust.TranslateBattle <- *update.CallbackQuery
				return true
			}
		}
	}

	// translate lang change
	if update.CallbackQuery != nil {
		ChangeLanguage(update.CallbackQuery)
		return true
	}

	// new game and message (continue)
	return false
}

func addToPlayerBase(PlayerID string, chatID int64) {
	cust.Players[PlayerID] = &cust.UsersStatistic{
		ChatID: chatID,
		Location: &cust.Localization{
			Language: "en",
		},
		Field: &cust.Field{
			Mutex:        true,
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	cust.SaveBase()
	cust.Players[PlayerID].ParseLangMap()
}

func NoneUserNamePlayer(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, but you don't have a Username in your telegram profile\nAdd it by going to Settings -> Edit -> Username")

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func StopGame(update *tgbotapi.Update) {
	if cust.Players[update.Message.From.UserName].RunGame {
		go game_logic.ListenCallbackQuery(update)
		cust.StopChannel <- *update.Message
	} else {
		msg := tgbotapi.NewMessage(cust.Players[update.Message.From.UserName].ChatID, cust.Players[update.Message.From.UserName].Location.Dictionary["norungame"])

		if _, err := cust.Bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
	fmt.Println("game stopped")
}

func SetLanguage(update *tgbotapi.Update) {
	playerID := update.Message.From.UserName
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, cust.Players[update.Message.From.UserName].Location.Dictionary["change_lang"])

	ru := tgbotapi.NewInlineKeyboardButtonData("Русский", "ru")
	en := tgbotapi.NewInlineKeyboardButtonData("English", "en")
	row := tgbotapi.NewInlineKeyboardRow(ru, en)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	msgData, err := cust.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	cust.Players[playerID].OccupiedSells = append(cust.Players[playerID].OccupiedSells, msgData.MessageID)
}

func ChangeLanguage(updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.UserName
	switch updateCallback.Data {
	case "ru":
		cust.Players[playerID].Location.Language = "ru"
	case "en":
		cust.Players[playerID].Location.Language = "en"
	default:
		msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["finished_game"])

		if _, err := cust.Bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}
	cust.Players[playerID].ParseLangMap()
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["lang_set"])

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}

	cust.SaveBase()
	go game_logic.DeleteMessage(playerID, cust.Players[playerID].ChatID)
}

func ChangeSide(update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	playerID := update.Message.From.UserName

	if cust.Players[playerID].FirstMove {
		cust.Players[playerID].FirstMove = false
		msg.Text = cust.Players[playerID].Location.Dictionary["play_zero"]
	} else {
		cust.Players[playerID].FirstMove = true
		msg.Text = cust.Players[playerID].Location.Dictionary["play_cross"]
	}

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
	cust.SaveBase()
}

func GameLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	playerID := update.Message.From.UserName
	game_logic.Tttgame(update)
	cust.Players[playerID].ChatID = update.Message.Chat.ID
	cust.Players[playerID].RunGame = true
	go game_logic.ListenCallbackQuery(update)

	cust.SaveBase()
}

func BattleLaunch(update *tgbotapi.Update) {
	var p2UserName string
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	p1UserName := update.Message.From.UserName
	if strings.Count(update.Message.Text, "/tttbattle @") != 1 {
		InvalidOpponentUsername(p1UserName)
		return
	}

	p2UserName = strings.Replace(update.Message.Text, "/tttbattle @", "", 1)
	if strings.Count(p2UserName, " ") > 0 || len(p2UserName) < 5 {
		InvalidOpponentUsername(p1UserName)
		return
	}

	fmt.Println(checkBattleBase(p1UserName, p2UserName))
}

func checkBattleBase(userName1, userName2 string) bool {
	if _, inBase := cust.Battles[userName1+","+userName2]; !inBase {
		addToBattlesBase(userName1, userName2)
		fmt.Println("Players added to battle map")
		return true
	}

	return true
}

func addToBattlesBase(userName1, userName2 string) {
	var firstMove bool
	switch rand.Intn(2) {
	case 1:
		firstMove = true
	case 2:
		firstMove = false
	}

	cust.Battles[userName1+","+userName2] = &cust.BattleStatistic{
		FirstMove: firstMove,
		Player1: &cust.Player{
			UserName: userName1,
			Queue:    firstMove,
		},
		Player2: &cust.Player{
			UserName: userName2,
			Queue:    !firstMove,
		},
		Field: &cust.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	cust.Battles[userName2+","+userName1] = &*cust.Battles[userName1+","+userName2]
	cust.Battles[userName1] = &*cust.Battles[userName1+","+userName2]
	cust.Battles[userName2] = &*cust.Battles[userName1+","+userName2]

	cust.SaveBase()
}

func InvalidOpponentUsername(playerID string) {
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["invalid_opponent"])

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func StartMsg(Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.UserName].Location.Dictionary["start"])

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func UnfinishedGameMsg(Message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, cust.Players[Message.From.UserName].Location.Dictionary["unfinished"])

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func RecognitionCommand(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "setlanguage":
		SetLanguage(update)
	case "changeside":
		ChangeSide(update)
	case "tttgame":
		GameLaunch(update)
	case "tttbattle":
		BattleLaunch(update)
	case "start":
		StartMsg(update.Message)
	default:
	}
}
