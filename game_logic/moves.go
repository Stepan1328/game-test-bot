package game_logic

import (
	cust "github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

var (
	crossButton = tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton  = tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")
)

func Tttgame(update *tgbotapi.Update) {
	playerID := update.Message.From.UserName
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["main"])
	msg.ReplyMarkup = ParseMarkUp(playerID)

	msgData, err := cust.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].MsgID = msgData.MessageID
	cust.SaveBase()
}

func ListenCallbackQuery(update *tgbotapi.Update) {
	if update.Message != nil {
		if !cust.Players[update.Message.From.UserName].FirstMove && cust.Players[update.Message.From.UserName].Field.Move == 1 {
			cust.Players[update.Message.From.UserName].BotMove()
		}
	} else {
		if !cust.Players[update.CallbackQuery.From.UserName].FirstMove && cust.Players[update.CallbackQuery.From.UserName].Field.Move == 1 {
			cust.Players[update.CallbackQuery.From.UserName].BotMove()
		}
	}

	go func() {
		for {
			if motion() {
				return
			}
		}
	}()
}

func motion() bool {
	select {
	case updateCallback := <-cust.TranslateUpdate:
		cust.Players[updateCallback.From.UserName].Field.Mutex = false
		if cust.Players[updateCallback.From.UserName].CheckMsg(updateCallback.Message.MessageID) {
			return makeDoubleMove(updateCallback)
		}

		TemporaryMessage(updateCallback.From.UserName, "irrelevant_field")
		return false

	case Message := <-cust.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(cust.Players[Message.From.UserName].ChatID, "Game stopped")

		if _, err := cust.Bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		cust.Players[Message.From.UserName].ClearField()
		cust.SaveBase()
		return true
	}
}

func makeDoubleMove(updateCallback tgbotapi.CallbackQuery) bool {
	//chatID := updateCallback.Message.Chat.ID
	playerID := updateCallback.From.UserName
	if updateCallback.Data == " " {
		TemporaryMessage(playerID, "occupied_cell")
		cust.SaveBase()
		return false
	}

	go DeleteMessage(playerID)

	if cust.Players[playerID].HumanMove(updateCallback.Data) {
		cust.Players[playerID].Field.Mutex = true
		return true
	}

	return cust.Players[playerID].BotMove()
}

func DeleteMessage(playerID string) {
	for len(cust.Players[playerID].OccupiedSells) > 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(cust.Players[playerID].ChatID, cust.Players[playerID].OccupiedSells[0])

		if _, err := cust.Bot.Send(deleteMsg); err != nil {
			log.Println(err)
		}

		cust.Players[playerID].OccupiedSells = cust.Players[playerID].OccupiedSells[1:]
	}
	cust.SaveBase()
}

func TemporaryMessage(playerID, text string) {
	replymsg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary[text])

	msgData, err := cust.Bot.Send(replymsg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].OccupiedSells = append(cust.Players[playerID].OccupiedSells, msgData.MessageID)
	cust.SaveBase()
}

func SimpleMsg(playerID, text string) {
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary[text])

	if _, err := cust.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func ParseMarkUp(playerID string) tgbotapi.InlineKeyboardMarkup {
	var masOfButton [9]tgbotapi.InlineKeyboardButton
	var masOfRow [3][]tgbotapi.InlineKeyboardButton

	for i := 1; i <= 9; i++ {
		switch cust.Players[playerID].Field.PlayingField[(i-1)/3][(i-1)%3] {
		case 0:
			button := tgbotapi.NewInlineKeyboardButtonData(" ", strconv.Itoa(i))
			masOfButton[i-1] = button
		case 1:
			masOfButton[i-1] = crossButton
		case 2:
			masOfButton[i-1] = zeroButton
		}
	}

	for i := 0; i < 3; i++ {
		buttonRow := tgbotapi.NewInlineKeyboardRow(masOfButton[i*3], masOfButton[i*3+1], masOfButton[i*3+2])
		masOfRow[i] = buttonRow
	}

	buttonMatrix := tgbotapi.NewInlineKeyboardMarkup(masOfRow[0], masOfRow[1], masOfRow[2])

	return buttonMatrix
}
