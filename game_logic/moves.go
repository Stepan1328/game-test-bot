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
	playerID := update.Message.From.ID
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["main"])
	msg.ReplyMarkup = parseMarkUp(playerID)

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	cust.Players[playerID].MsgID = msgData.MessageID
}

func ListenCallbackQuery(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if !cust.Players[update.Message.From.ID].FirstMove {
		botMove(bot, update.Message.From.ID)
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
		if cust.Players[updateCallback.From.ID].CheckMsg(updateCallback.Message.MessageID) {
			return makeDoubleMove(bot, updateCallback)
		}

		irrelevantField(bot, updateCallback.From.ID)
		return false

	case Message := <-cust.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(int64(Message.From.ID), "Game stopped")

		if _, err := bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		cust.Players[Message.From.ID].ClearField()

		return true
	}
}

func makeDoubleMove(bot *tgbotapi.BotAPI, updateCallback tgbotapi.CallbackQuery) bool {
	chatID := updateCallback.Message.Chat.ID
	if updateCallback.Data == " " {
		occupiedSell(bot, chatID, updateCallback.From.ID)
		return false
	}

	go DeleteMessage(bot, chatID)

	if humanMove(bot, updateCallback) {
		return true
	}

	return botMove(bot, updateCallback.From.ID)
}

func occupiedSell(bot *tgbotapi.BotAPI, chatID int64, playerID int) {
	replymsg := tgbotapi.NewMessage(chatID, cust.Players[playerID].Location.Dictionary["occupied_cell"])

	msgData, err := bot.Send(replymsg)
	if err != nil {
		log.Println(err)
	}

	cust.OccupiedSells = append(cust.OccupiedSells, msgData.MessageID)
}

func humanMove(bot *tgbotapi.BotAPI, updateCallback tgbotapi.CallbackQuery) bool {
	playerID := updateCallback.From.ID
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

func botMove(bot *tgbotapi.BotAPI, playerID int) bool {
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

func checkSituation(bot *tgbotapi.BotAPI, playerID int) bool {
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

func ParseLangMap(playerID int) {
	lang := cust.Players[playerID].Location.Language
	bytes, _ := os.ReadFile("./assets/" + lang + ".json")

	_ = json.Unmarshal(bytes, &cust.Players[playerID].Location.Dictionary)
}

func DeleteMessage(bot *tgbotapi.BotAPI, chatID int64) {
	for len(cust.OccupiedSells) > 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, cust.OccupiedSells[0])

		if _, err := bot.Send(deleteMsg); err != nil {
			log.Println(err)
		}

		cust.OccupiedSells = cust.OccupiedSells[1:]
	}
}

func irrelevantField(bot *tgbotapi.BotAPI, playerID int) {
	msg := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["irrelevant_field"])

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	cust.OccupiedSells = append(cust.OccupiedSells, msgData.MessageID)
}

func sendDrawMsg(bot *tgbotapi.BotAPI, playerID int) {
	drawMessage := tgbotapi.NewMessage(cust.Players[playerID].ChatID, cust.Players[playerID].Location.Dictionary["draw"])

	if _, err := bot.Send(drawMessage); err != nil {
		log.Println(err)
	}

	cust.Players[playerID].ClearField()
}

func sendWinMsg(bot *tgbotapi.BotAPI, playerID int) bool {
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

func parseMarkUp(playerID int) tgbotapi.InlineKeyboardMarkup {
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
