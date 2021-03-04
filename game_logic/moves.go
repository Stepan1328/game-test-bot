package game_logic

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

var (
	crossButton = tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton  = tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")
)

func Tttgame(update tgbotapi.Update, bot tgbotapi.BotAPI) int {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "This is a tic-tac-toe field\n"+
		"To put a cross or a zero, just click on the button and its state will change\nGood luck")

	msg.ReplyMarkup = parseMarkUp(update.Message.From.ID)

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	return msgData.MessageID
}

func ListenCallbackQuery(update tgbotapi.Update, bot tgbotapi.BotAPI, msgID int) {
	move := 1

	situation := Situation{PlayField: cust.Players[update.Message.From.ID].PlayingField}
	motion, _ := situation.Analyze((move+1)%2+1, move)

	cust.Players[update.Message.From.ID].PlayingField[motion.Y][motion.X] = (move+1)%2 + 1
	buttonMatrix := parseMarkUp(update.Message.From.ID)
	move++

	replymsg := tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID, msgID, buttonMatrix)

	if _, err := bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	go func() {
		for {
			if MakeDoubleMove(bot, &move) {
				return
			}
		}
	}()
}

func MakeDoubleMove(bot tgbotapi.BotAPI, move *int) bool {
	select {
	case updateCallback := <-cust.TranslateUpdate:
		chatID := updateCallback.Message.Chat.ID
		playerID := updateCallback.From.ID
		if updateCallback.Data == " " {
			fmt.Println("the occupied cell is pressed")

			replymsg := tgbotapi.NewMessage(chatID, "This cell is already occupied\nSelect another cell")

			msgData, err := bot.Send(replymsg)
			if err != nil {
				log.Println(err)
			}

			cust.OccupiedSells = append(cust.OccupiedSells, msgData.MessageID)

			return false
		}

		for len(cust.OccupiedSells) > 0 {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, cust.OccupiedSells[0])

			if _, err := bot.Send(deleteMsg); err != nil {
				log.Println(err)
			}

			cust.OccupiedSells = cust.OccupiedSells[1:]
			fmt.Println("message deleted")
		}

		numberOfCell, err := strconv.Atoi(updateCallback.Data)
		if err != nil {
			log.Println(err)
		}

		column := (numberOfCell - 1) % 3
		row := (numberOfCell - 1) / 3

		cust.Players[playerID].PlayingField[row][column] = (*move+1)%2 + 1
		buttonMatrix := parseMarkUp(playerID)
		*move++

		replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

		if _, err = bot.Send(replymsg); err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, playerID, *move) {
			cust.Players[playerID].ClearField()
			return true
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			if _, err = bot.Send(drawMessage); err != nil {
				log.Println(err)
			}

			cust.Players[playerID].ClearField()

			return true
		}

		situation := Situation{PlayField: cust.Players[playerID].PlayingField}
		motion, _ := situation.Analyze((*move+1)%2+1, *move)

		cust.Players[playerID].PlayingField[motion.Y][motion.X] = (*move+1)%2 + 1
		buttonMatrix = parseMarkUp(playerID)
		*move++

		replymsg = tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)
		time.Sleep(time.Millisecond * 300)

		_, err = bot.Send(replymsg)
		if err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, playerID, *move) {
			cust.Players[playerID].ClearField()
			return true
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			if _, err = bot.Send(drawMessage); err != nil {
				log.Println(err)
			}

			cust.Players[playerID].ClearField()
			fmt.Println("Draw")

			return true
		}

	case Message := <-cust.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(int64(Message.From.ID), "Game stopped")

		if _, err := bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		cust.Players[Message.From.ID].RunGame = false
		cust.Players[Message.From.ID].PlayingField = [3][3]int{}

		return true
	}

	return false
}

func sendWinMsg(bot tgbotapi.BotAPI, chatID int64, playerID int, move int) bool {
	win, _ := CheckingWinner(cust.Players[playerID].PlayingField, move-1)

	if !win {
		return win
	}

	winMessage := tgbotapi.NewMessage(chatID, "")

	if move%2 == 0 {
		winMessage.Text = "Congratulations to 🔥Player 1🔥 he is the winner 🎉\n" +
			"Player 2 don't get upset and ask for a rematch by writing /tttgame"
	} else {
		winMessage.Text = "Congratulations to 🔥Player 2🔥 he is the winner 🎉\n" +
			"Player 1 don't get upset and ask for a rematch by writing /tttgame"
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
		switch cust.Players[playerID].PlayingField[(i-1)/3][(i-1)%3] {
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
