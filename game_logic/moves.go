package game_logic

import (
	"encoding/json"
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"strconv"
)

var (
	crossButton = tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton  = tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")
)

func Tttgame(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	playerID := update.Message.From.UserName
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["main"])
	msg.ReplyMarkup = parseMarkUp(playerID)

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].MsgID = msgData.MessageID
}

func ListenCallbackQuery(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if !cust.Players[update.Message.From.UserName].FirstMove {
		botMove(bot, update.Message.From.UserName)
	}

	go func() {
		for {
			if motion(bot) {
				return
			}
		}
	}()
}

func motion(bot *tgbotapi.BotAPI) bool {
	select {
	case updateCallback := <-cust.TranslateUpdate:
		if cust.Players[updateCallback.From.UserName].CheckMsg(updateCallback.Message.MessageID) {
			return makeDoubleMove(bot, updateCallback)
		}

		irrelevantField(bot, updateCallback.From.UserName)
		return false

	case Message := <-cust.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(int64(Message.From.ID), "Game stopped")

		if _, err := bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		cust.Players[Message.From.UserName].ClearField()

		return true
	}
}

func makeDoubleMove(bot *tgbotapi.BotAPI, updateCallback tgbotapi.CallbackQuery) bool {
	chatID := updateCallback.Message.Chat.ID
	if updateCallback.Data == " " {
		occupiedSell(bot, chatID, updateCallback.From.UserName)
		return false
	}

	go DeleteMessage(bot, updateCallback.From.UserName, chatID)

	if humanMove(bot, updateCallback) {
		return true
	}

	return botMove(bot, updateCallback.From.UserName)
}

func occupiedSell(bot *tgbotapi.BotAPI, chatID int64, playerID string) {
	replymsg := tgbotapi.NewMessage(chatID, cust.Players[playerID].Location.Dictionary["occupied_cell"])

	msgData, err := bot.Send(replymsg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].OccupiedSells = append(cust.Players[playerID].OccupiedSells, msgData.MessageID)
}

func humanMove(bot *tgbotapi.BotAPI, updateCallback tgbotapi.CallbackQuery) bool {
	playerID := updateCallback.From.UserName
	move := cust.Players[playerID].Field.Move
	chatID := cust.Players[playerID].ChatID
	numberOfCell, err := strconv.Atoi(updateCallback.Data)
	if err != nil {
		log.Println(err)
	}

	column := (numberOfCell - 1) % 3
	row := (numberOfCell - 1) / 3
	cust.Players[playerID].Field.PlayingField[row][column] = (move+1)%2 + 1
	cust.Players[playerID].Field.Move += 1
	buttonMatrix := parseMarkUp(playerID)

	replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

	if _, err = bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	if checkSituation(bot, playerID) {
		return true
	}

	return false
}

func botMove(bot *tgbotapi.BotAPI, playerID string) bool {
	move := cust.Players[playerID].Field.Move
	situation := Situation{PlayField: cust.Players[playerID].Field.PlayingField}
	motion, _ := situation.Analyze((move+1)%2+1, move)
	cust.Players[playerID].Field.PlayingField[motion.Y][motion.X] = (move+1)%2 + 1
	cust.Players[playerID].Field.Move += 1

	buttonMatrix := parseMarkUp(playerID)

	replymsg := tgbotapi.NewEditMessageReplyMarkup(cust.Players[playerID].ChatID, cust.Players[playerID].MsgID, buttonMatrix)

	if _, err := bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	if checkSituation(bot, playerID) {
		return true
	}

	return false
}

func checkSituation(bot *tgbotapi.BotAPI, playerID string) bool {
	move := cust.Players[playerID].Field.Move

	if sendWinMsg(bot, playerID) {
		cust.Players[playerID].ClearField()
		return true
	}

	if move > 9 {
		go sendDrawMsg(bot, playerID)
		return true
	}

	return false
}

func ParseLangMap(playerID string) {
	lang := cust.Players[playerID].Location.Language
	bytes, _ := os.ReadFile("./assets/" + lang + ".json")

	_ = json.Unmarshal(bytes, &cust.Players[playerID].Location.Dictionary)
}

func DeleteMessage(bot *tgbotapi.BotAPI, playerID string, chatID int64) {
	for len(cust.Players[playerID].OccupiedSells) > 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, cust.Players[playerID].OccupiedSells[0])

		if _, err := bot.Send(deleteMsg); err != nil {
			log.Println(err)
		}

		cust.Players[playerID].OccupiedSells = cust.Players[playerID].OccupiedSells[1:]
	}
}

func irrelevantField(bot *tgbotapi.BotAPI, playerID string) {
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["irrelevant_field"])

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].OccupiedSells = append(cust.Players[playerID].OccupiedSells, msgData.MessageID)
}

func sendDrawMsg(bot *tgbotapi.BotAPI, playerID string) {
	drawMessage := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["draw"])

	if _, err := bot.Send(drawMessage); err != nil {
		log.Println(err)
	}

	cust.Players[playerID].ClearField()
}

func sendWinMsg(bot *tgbotapi.BotAPI, playerID string) bool {
	move := cust.Players[playerID].Field.Move
	win, _ := CheckingWinner(cust.Players[playerID].Field.PlayingField, move-1)

	if !win {
		return win
	}

	winMessage := tgbotapi.NewMessage(cust.Players[playerID].ChatID, "")

	if move%2 == 0 {
		winMessage.Text = cust.Players[playerID].Location.Dictionary["win_cross"]
	} else {
		winMessage.Text = cust.Players[playerID].Location.Dictionary["win_zero"]
	}

	if _, err := bot.Send(winMessage); err != nil {
		log.Println(err)
	}

	fmt.Println("found the winner")

	return win
}

func parseMarkUp(playerID string) tgbotapi.InlineKeyboardMarkup {
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
