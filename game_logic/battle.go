package gamelogic

import (
	"fmt"
	"github.com/Stepan1328/game-test-bot/clients"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"sync"
)

// Tttbattle is a function that sends a request to the opponent for the game
func Tttbattle(userName string) { // TODO: request send msg
	u1, u2 := identifyPlayers(userName)

	text := fmt.Sprintf(clients.Players[u2].Location.Dictionary["battle_req"], clients.Players[u1].UserName)
	msg := tgbotapi.NewMessage(clients.Players[u2].ChatID, text)
	yes := tgbotapi.NewInlineKeyboardButtonData(clients.Players[u2].Location.Dictionary["yes"], "yes")
	no := tgbotapi.NewInlineKeyboardButtonData(clients.Players[u2].Location.Dictionary["no"], "no")
	row := tgbotapi.NewInlineKeyboardRow(yes, no)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

	clients.Players[u2].InvitationStack = append(clients.Players[u2].InvitationStack, msg)
	userName2 := clients.Players[u2].UserName

	if !clients.Players[u2].RunGame && !clients.Battles[userName2].RunGame {
		clients.CheckInvitationStack(u2)
	}
}

// AnalyzeResponseToRequest is a function that analyzes the response to a request
func AnalyzeResponseToRequest(callback *tgbotapi.CallbackQuery) {
	userName := callback.From.UserName
	clients.Battles[userName].ParseMarkUp()
	uID1, uID2 := identifyPlayers(userName)

	switch callback.Data {
	case "yes":
		uN1 := clients.Players[uID1].UserName
		uN2 := clients.Players[uID2].UserName
		if clients.Players[uID2].RunGame || clients.Battles[uN2].RunGame {
			TemporaryMessage(uID1, "player_is_busy")
			clients.SaveBase()
			return
		}
		DeleteMessage(uID1)

		SimpleMsg(uID1, "accepted_invitation")
		SimpleMsg(uID2, "accepted_req")
		clients.Battles[uN1] = &*clients.Battles[uN2+","+uN1]
		clients.Battles[uN2] = &*clients.Battles[uN2+","+uN1]
		clients.Players[uID1].LastBattleID = uID2
		clients.Players[uID2].LastBattleID = uID1

		sendDouMsg(uN2)
		clients.SaveBase()
	case "no":
		DeleteMessage(uID1)
		SimpleMsg(uID1, "declined_invitation")
		SimpleMsg(uID2, "declined_req")

		clients.CheckInvitationStack(uID1)
	}

	clients.SaveBase()
}

func sendDouMsg(userName string) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		clients.Battles[userName].Player1.MsgID = sendMsg(userName, 1).MessageID

		wg.Done()
	}()
	go func() {
		clients.Battles[userName].Player2.MsgID = sendMsg(userName, 2).MessageID
		wg.Done()
	}()
	wg.Wait()
}

func sendMsg(userName string, numOfPlayer int) tgbotapi.Message {
	var chatID int64
	var text string

	switch numOfPlayer {
	case 1:
		u1 := clients.Battles[userName].Player1.PlayerID
		chatID = clients.Players[u1].ChatID
		if clients.Battles[clients.Players[u1].UserName].FirstMove {
			text = clients.Players[u1].Location.Dictionary["battle_cross_his_turn"]
		} else {
			text = clients.Players[u1].Location.Dictionary["battle_zero_no_his_turn"]
		}
	case 2:
		u2 := clients.Battles[userName].Player2.PlayerID
		chatID = clients.Players[u2].ChatID
		if clients.Battles[clients.Players[u2].UserName].FirstMove {
			text = clients.Players[u2].Location.Dictionary["battle_zero_no_his_turn"]
		} else {
			text = clients.Players[u2].Location.Dictionary["battle_cross_his_turn"]
		}
	}

	clients.Battles[userName].RunGame = true
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = clients.Battles[userName].FieldMarkup

	msgData, err := clients.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return msgData
}

func identifyPlayers(userName string) (int, int) {
	var u1, u2 int

	if clients.Battles[userName].Player1.UserName == userName {
		u1 = clients.Compatibility[userName]
		u2 = clients.Battles[userName].Player2.PlayerID
	} else {
		u1 = clients.Compatibility[userName]
		u2 = clients.Battles[userName].Player1.PlayerID
	}

	return u1, u2
}

// BattleMotion is a function that processes callback responses
func BattleMotion() {
	select {
	case updateCallback := <-clients.TranslateBattle:
		userName := updateCallback.From.UserName
		msgID := updateCallback.Message.MessageID
		if !clients.Battles[userName].CheckMsg(msgID, userName) {
			TemporaryMessage(updateCallback.From.ID, "irrelevant_field")
			return
		}

		makeMove(updateCallback)
	case Message := <-clients.StopBattleChannel:
		stopGameMsg(clients.Battles[Message.From.UserName].Player1.PlayerID)
		stopGameMsg(clients.Battles[Message.From.UserName].Player2.PlayerID)

		clients.Battles[Message.From.UserName].ClearField()
		clients.SaveBase()

		clients.CheckInvitationStack(clients.Compatibility[Message.From.UserName])
	}
}

func makeMove(updateCallback tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.ID
	userName := updateCallback.From.UserName
	if !clients.Battles[userName].CheckQueue(userName) {
		TemporaryMessage(playerID, "not_player_move")
		clients.SaveBase()
		return
	}

	if updateCallback.Data == " " {
		TemporaryMessage(playerID, "occupied_cell")
		clients.SaveBase()
		return
	}

	clients.Battles[updateCallback.From.UserName].HumanBattleMove(updateCallback.Data)
	DeleteMessage(clients.Battles[userName].Player1.PlayerID)
	DeleteMessage(clients.Battles[userName].Player2.PlayerID)
}

func stopGameMsg(playerID int) {
	stopGameMessage := tgbotapi.NewMessage(clients.Players[playerID].ChatID, "Game stopped")

	if _, err := clients.Bot.Send(stopGameMessage); err != nil {
		log.Println(err)
	}
}
