package game_logic

import (
	"github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"sync"
)

func Tttbattle(userName1, userName2 string) {
	sendDouMsg(userName1, userName2, "battle_start")
	cust.SaveBase()
}

func sendDouMsg(userName1, userName2 string, msgText string) {
	wg := new(sync.WaitGroup)
	gluedUsername := userName1 + "," + userName2
	wg.Add(2)
	go func() {
		cust.Battles[gluedUsername].Player1.MsgID = sendMsg(userName1, userName2, msgText).MessageID
		wg.Done()
	}()
	go func() {
		cust.Battles[gluedUsername].Player2.MsgID = sendMsg(userName1, userName2, msgText).MessageID
		wg.Done()
	}()
	wg.Wait()
}

func sendMsg(userName1, userName2, msgText string) tgbotapi.Message {
	chatID := cust.Players[userName1].ChatID
	text := cust.Players[userName1].Location.Dictionary[msgText]

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = parseBattleMarkUp(userName1, userName2)

	msgData, err := cust.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	return msgData
}

func parseBattleMarkUp(userName1, userName2 string) tgbotapi.InlineKeyboardMarkup {
	var masOfButton [9]tgbotapi.InlineKeyboardButton
	var masOfRow [3][]tgbotapi.InlineKeyboardButton
	gluedUsername := userName1 + "," + userName2

	for i := 1; i <= 9; i++ {
		switch cust.Battles[gluedUsername].Field.PlayingField[(i-1)/3][(i-1)%3] {
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
