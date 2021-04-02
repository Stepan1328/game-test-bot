package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/Stepan1328/game-test-bot/clients"
	gl "github.com/Stepan1328/game-test-bot/game_logic"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func actionsWithUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		checkUpdate(&update)
	}
}

func checkUpdate(update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if checkPlayer(update) {
		return
	}

	fmt.Println(update.Message.Text)

	if update.Message.Command() == "stopgame" {
		stopGame(update)
		return
	}

	runGame := clients.Players[update.Message.From.ID].RunGame
	runBattle := false
	if _, inBase := clients.Battles[update.Message.From.UserName]; inBase {
		runBattle = clients.Battles[update.Message.From.UserName].RunGame
	}

	if runGame {
		gl.SimpleMsg(update.Message.From.ID, "unfinished_game")
		return
	}

	if _, inBase := clients.Battles[update.Message.From.UserName]; inBase {
		if update.Message.Command() == "leave" {
			stopBattle(update)
			return
		}
	}

	if runBattle {
		gl.SimpleMsg(update.Message.From.ID, "unfinished_battle")
		return
	}

	if update.Message.Command() != "" && !runGame && !runBattle {
		recognitionCommand(update)
		return
	}
}

func checkPlayer(update *tgbotapi.Update) bool {
	// check username and is it in the base via Message
	if update.Message != nil {
		if _, inBase := clients.Players[update.Message.From.ID]; !inBase {
			addToPlayerBase(update.Message.From.ID, update.Message.Chat.ID)
			fmt.Println("PLayer added to player base")
		}

		if update.Message.From.UserName != "" {
			clients.Players[update.Message.From.ID].UserName = update.Message.From.UserName
			clients.Compatibility[update.Message.From.UserName] = update.Message.From.ID
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
	callbackQuery := update.CallbackQuery

	// check username and is it in the base via Callback
	if _, inBase := clients.Players[callbackQuery.From.ID]; !inBase {
		addToPlayerBase(callbackQuery.From.ID, callbackQuery.Message.Chat.ID)
		fmt.Println("PLayer added to player base")
	}

	if callbackQuery.From.UserName != "" {
		clients.Players[callbackQuery.From.ID].UserName = callbackQuery.From.UserName
		clients.Compatibility[callbackQuery.From.UserName] = callbackQuery.From.ID
	}

	// translate gameUpdate (check run game and check Unlock mutex)
	if clients.Players[callbackQuery.From.ID].RunGame {
		clients.TranslateUpdate <- *update.CallbackQuery
		return
	}

	// translate battleUpdate (check run game and check Unlock mutex)
	if _, inBase := clients.Battles[callbackQuery.From.UserName]; inBase {
		if clients.Battles[callbackQuery.From.UserName].RunGame {
			clients.TranslateBattle <- *callbackQuery
			return
		}
	}

	// analyzing the response to a request
	if callbackQuery.Data == "yes" || callbackQuery.Data == "no" {
		gl.AnalyzeResponseToRequest(callbackQuery)
		return
	}

	// translate lang change
	changeLanguage(callbackQuery)
}

func addToPlayerBase(PlayerID int, chatID int64) {
	clients.Players[PlayerID] = &clients.UsersStatistic{
		BattleInvite: true,
		ChatID:       chatID,
		Location: &clients.Localization{
			Language: "en",
		},
		Field: &clients.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	clients.SaveBase()
	clients.Players[PlayerID].ParseLangMap()
}

func noneUserNamePlayer(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, but you don't have a Username in your telegram profile\nAdd it by going to Settings -> Edit -> Username")

	if _, err := clients.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func stopGame(update *tgbotapi.Update) {
	if clients.Players[update.Message.From.ID].RunGame {
		clients.StopChannel <- *update.Message
		fmt.Println("game stopped")
	} else {
		gl.SimpleMsg(update.Message.From.ID, "no_run_game")
		fmt.Println("no games started")
	}
}

func stopBattle(update *tgbotapi.Update) {
	if clients.Battles[update.Message.From.UserName].RunGame {
		clients.StopBattleChannel <- *update.Message
		fmt.Println("battle stopped")
	} else {
		gl.SimpleMsg(update.Message.From.ID, "no_run_battle")
		fmt.Println("no battles started")
	}
}

func setLanguage(update *tgbotapi.Update) {
	playerID := update.Message.From.ID
	msg := tgbotapi.NewMessage(clients.Players[playerID].ChatID, clients.Players[playerID].Location.Dictionary["change_lang"])

	ru := tgbotapi.NewInlineKeyboardButtonData("Русский", "ru")
	en := tgbotapi.NewInlineKeyboardButtonData("English", "en")
	row := tgbotapi.NewInlineKeyboardRow(ru, en)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	msgData, err := clients.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	clients.Players[playerID].OccupiedSells = append(clients.Players[playerID].OccupiedSells, msgData.MessageID)
}

func changeSide(update *tgbotapi.Update) {
	playerID := update.Message.From.ID

	if clients.Players[playerID].FirstMove {
		clients.Players[playerID].FirstMove = false
		gl.SimpleMsg(playerID, "play_zero")
	} else {
		clients.Players[playerID].FirstMove = true
		gl.SimpleMsg(playerID, "play_cross")
	}
	clients.SaveBase()
}

func gameLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	playerID := update.Message.From.ID
	gl.Tttgame(update)
	clients.Players[playerID].ChatID = update.Message.Chat.ID
	clients.Players[playerID].RunGame = true

	gl.FirstMove(update)
	clients.SaveBase()
}

func battleLaunch(update *tgbotapi.Update) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	userName1 := update.Message.From.UserName
	if userName1 == "" {
		noneUserNamePlayer(update.Message.Chat.ID)
		return
	}

	userID2, okBase := checkBattleBase(update)
	if !okBase {
		return
	}

	if clients.Players[userID2].BattleInvite {
		gl.Tttbattle(userName1)
	} else {
		gl.SimpleMsg(update.Message.From.ID, "player_dis_notify")
		return
	}
}

func checkBattleBase(update *tgbotapi.Update) (int, bool) {
	userID1 := update.Message.From.ID
	if strings.Count(update.Message.Text, "/tttbattle @") != 1 {
		gl.SimpleMsg(userID1, "invalid_opponent")
		return 0, false
	}

	userName2 := strings.Replace(update.Message.Text, "/tttbattle @", "", 1)
	if strings.Count(userName2, " ") > 0 || len(userName2) < 5 {
		gl.SimpleMsg(userID1, "invalid_opponent")
		return 0, false
	}

	userName1 := clients.Players[userID1].UserName
	if userName1 == userName2 {
		gl.SimpleMsg(userID1, "reversible_req")
		return 0, false
	}

	userID2 := clients.Compatibility[userName2]

	if _, inBase := clients.Players[userID2]; !inBase {
		gl.SimpleMsg(userID1, "player_not_found")
		return 0, false
	}

	if _, inBase := clients.Battles[userName1+","+userName2]; !inBase {
		addToBattlesBase(userName1, userName2)
		fmt.Println("Players added to battle map")
		return userID2, true
	}

	clients.Battles[userName1] = &*clients.Battles[userName1+","+userName2]
	clients.Battles[userName2] = &*clients.Battles[userName1+","+userName2]

	return userID2, true
}

func reMatch(update *tgbotapi.Update) {
	userID1 := update.Message.From.ID
	userID2 := clients.Players[userID1].LastBattleID
	if userID2 == 0 {
		gl.SimpleMsg(userID1, "no_battles_played")
		return
	}
	userName1 := clients.Battles[update.Message.From.UserName].Player1.UserName
	userName2 := clients.Battles[update.Message.From.UserName].Player2.UserName

	clients.Battles[userName1] = &*clients.Battles[userName1+","+userName2]
	clients.Battles[userName2] = &*clients.Battles[userName1+","+userName2]

	if clients.Players[userID2].BattleInvite {
		gl.Tttbattle(update.Message.From.UserName)
	} else {
		gl.SimpleMsg(update.Message.From.ID, "player_dis_notify")
		return
	}
}

func addToBattlesBase(userName1, userName2 string) {
	var firstMove bool
	switch rand.Intn(2) {
	case 1:
		firstMove = true
	case 2:
		firstMove = false
	}

	clients.Battles[userName1+","+userName2] = &clients.BattleStatistic{
		FirstMove: firstMove,
		Player1: &clients.Player{
			UserName: userName1,
			PlayerID: clients.Compatibility[userName1],
			Queue:    firstMove,
		},
		Player2: &clients.Player{
			UserName: userName2,
			PlayerID: clients.Compatibility[userName2],
			Queue:    !firstMove,
		},
		Field: &clients.Field{
			PlayingField: [3][3]int{},
			Move:         1,
		},
	}

	clients.Battles[userName2+","+userName1] = &*clients.Battles[userName1+","+userName2]
	clients.Battles[userName1] = &*clients.Battles[userName1+","+userName2]
	clients.Battles[userName2] = &*clients.Battles[userName1+","+userName2]

	clients.SaveBase()
}

func changeBattleInvite(update *tgbotapi.Update) {
	playerID := update.Message.From.ID

	if clients.Players[playerID].BattleInvite {
		clients.Players[playerID].BattleInvite = false
		gl.SimpleMsg(playerID, "disable_notifications")
	} else {
		clients.Players[playerID].BattleInvite = true
		gl.SimpleMsg(playerID, "enable_notifications")
	}

	clients.SaveBase()
}

func changeLanguage(updateCallback *tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.ID
	switch updateCallback.Data {
	case "ru":
		clients.Players[playerID].Location.Language = "ru"
	case "en":
		clients.Players[playerID].Location.Language = "en"
	default:
		gl.SimpleMsg(playerID, "finished_game")
		return
	}
	clients.Players[playerID].ParseLangMap()

	gl.SimpleMsg(playerID, "lang_set")

	clients.SaveBase()
	go gl.DeleteMessage(playerID)
}

func recognitionCommand(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "setlanguage":
		setLanguage(update)
	case "changeside":
		changeSide(update)
	case "tttgame":
		gameLaunch(update)
	case "tttbattle":
		battleLaunch(update)
	case "muteinvite":
		changeBattleInvite(update)
	case "rematch":
		reMatch(update)
	case "start":
		gl.SimpleMsg(update.Message.From.ID, "start")
	default:
	}
}
