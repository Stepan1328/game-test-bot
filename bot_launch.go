package main

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	gl "github.com/Stepan1328/game-test-bot/game_logic"
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
		gl.SimpleMsg(Message.From.UserName, "unfinished")
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

	// check update.CallbackQuery
	if update.CallbackQuery != nil {
		checkCallback(update)
		return true
	}

	// new game and message (continue)
	return false
}

func checkCallback(update *tgbotapi.Update) {
	CallbackQuery := update.CallbackQuery

	// check username and is it in the base via Callback
	if CallbackQuery.From.UserName == "" {
		NoneUserNamePlayer(CallbackQuery.Message.Chat.ID)
		fmt.Println("NoneUserNamePlayer")
		return
	}

	if _, inBase := cust.Players[CallbackQuery.From.UserName]; !inBase {
		addToPlayerBase(CallbackQuery.From.UserName, CallbackQuery.Message.Chat.ID)
		fmt.Println("PLayer added to player base")
	}

	// translate gameUpdate (check run game and check Unlock mutex)
	if cust.Players[CallbackQuery.From.UserName].RunGame {
		if !cust.Players[CallbackQuery.From.UserName].Field.Mutex {
			return
		}
		go gl.ListenCallbackQuery(update)
		cust.TranslateUpdate <- *update.CallbackQuery
		return
	}

	// translate battleUpdate (check run game and check Unlock mutex)
	if _, inBase := cust.Battles[CallbackQuery.From.UserName]; inBase {
		if cust.Players[CallbackQuery.From.UserName].RunGame { // TODO: mutex
			cust.TranslateBattle <- *CallbackQuery // TODO: checking the opponent's employment
		}
	}

	// translate lang change
	ChangeLanguage(CallbackQuery)
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
		go gl.ListenCallbackQuery(update)
		cust.StopChannel <- *update.Message
	} else {
		gl.SimpleMsg(update.Message.From.UserName, "norungame")
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

func ChangeSide(update *tgbotapi.Update) {
	playerID := update.Message.From.UserName

	if cust.Players[playerID].FirstMove {
		cust.Players[playerID].FirstMove = false
		gl.SimpleMsg(playerID, "play_zero")
	} else {
		cust.Players[playerID].FirstMove = true
		gl.SimpleMsg(playerID, "play_cross")
	}
	cust.SaveBase()
}

func GameLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	playerID := update.Message.From.UserName
	gl.Tttgame(update)
	cust.Players[playerID].ChatID = update.Message.Chat.ID
	cust.Players[playerID].RunGame = true
	go gl.ListenCallbackQuery(update)

	cust.SaveBase()
}

func BattleLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	if !checkBattleBase(update) {
		return
	}

	go gl.Tttbattle(update.Message.From.UserName)
}

func checkBattleBase(update *tgbotapi.Update) bool {
	p1UserName := update.Message.From.UserName
	if strings.Count(update.Message.Text, "/tttbattle @") != 1 {
		gl.SimpleMsg(p1UserName, "invalid_opponent")
		return false
	}

	p2UserName := strings.Replace(update.Message.Text, "/tttbattle @", "", 1)
	if strings.Count(p2UserName, " ") > 0 || len(p2UserName) < 5 {
		gl.SimpleMsg(p1UserName, "invalid_opponent")
		return false
	}

	if p1UserName == p2UserName {
		gl.SimpleMsg(p1UserName, "reversible_req")
		return false
	}

	if _, inBase := cust.Battles[p1UserName+","+p2UserName]; !inBase {
		addToBattlesBase(p1UserName, p2UserName)
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

func ChangeLanguage(updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.UserName
	switch updateCallback.Data {
	case "ru":
		cust.Players[playerID].Location.Language = "ru"
	case "en":
		cust.Players[playerID].Location.Language = "en"
	default:
		gl.SimpleMsg(playerID, "finished_game")
		return
	}
	cust.Players[playerID].ParseLangMap()

	gl.SimpleMsg(playerID, "lang_set")

	cust.SaveBase()
	go gl.DeleteMessage(playerID)
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
		gl.SimpleMsg(update.Message.From.UserName, "start")
	default:
	}
}
