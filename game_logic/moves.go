package game_logic

import (
	"fmt"
	cust "github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

func Tttgame(update tgbotapi.Update, bot tgbotapi.BotAPI) (tgbotapi.InlineKeyboardMarkup, int) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "This is a tic-tac-toe field\n"+
		"To put a cross or a zero, just click on the button and its state will change\nGood luck")

	button1 := tgbotapi.NewInlineKeyboardButtonData(" ", "1")
	button2 := tgbotapi.NewInlineKeyboardButtonData(" ", "2")
	button3 := tgbotapi.NewInlineKeyboardButtonData(" ", "3")
	button4 := tgbotapi.NewInlineKeyboardButtonData(" ", "4")
	button5 := tgbotapi.NewInlineKeyboardButtonData(" ", "5")
	button6 := tgbotapi.NewInlineKeyboardButtonData(" ", "6")
	button7 := tgbotapi.NewInlineKeyboardButtonData(" ", "7")
	button8 := tgbotapi.NewInlineKeyboardButtonData(" ", "8")
	button9 := tgbotapi.NewInlineKeyboardButtonData(" ", "9")

	buttonRow1 := tgbotapi.NewInlineKeyboardRow(button1, button2, button3)
	buttonRow2 := tgbotapi.NewInlineKeyboardRow(button4, button5, button6)
	buttonRow3 := tgbotapi.NewInlineKeyboardRow(button7, button8, button9)

	buttonMatrix := tgbotapi.NewInlineKeyboardMarkup(buttonRow1, buttonRow2, buttonRow3)

	msg.ReplyMarkup = buttonMatrix

	msgData, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	return buttonMatrix, msgData.MessageID
}

func ListenCallbackQuery(update tgbotapi.Update, bot tgbotapi.BotAPI, buttonMatrix tgbotapi.InlineKeyboardMarkup, msgID int) {
	move := 1

	for i := 0; i < 3; i++ {
		fmt.Println(cust.Players[update.Message.From.ID].PlayingField[i])
	}
	fmt.Println("==")
	cust.Players[update.Message.From.ID].PlayingField = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	for i := 0; i < 3; i++ {
		fmt.Println(cust.Players[update.Message.From.ID].PlayingField[i])
	}
	fmt.Println("==")

	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")

	situation := Situation{PlayField: cust.Players[update.Message.From.ID].PlayingField}
	motion, _ := situation.Analyze((move+1)%2+1, move)

	buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
	cust.Players[update.Message.From.ID].PlayingField[motion.Y][motion.X] = (move+1)%2 + 1
	move++
	for i := 0; i < 3; i++ {
		fmt.Println(cust.Players[update.Message.From.ID].PlayingField[i])
	}

	replymsg := tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID, msgID, buttonMatrix)

	time.Sleep(time.Millisecond * 200)

	if _, err := bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	go func() {
		for {
			MakeDoubleMove(bot, buttonMatrix, &move)
		}
	}()
}

func MakeDoubleMove(bot tgbotapi.BotAPI, buttonMatrix tgbotapi.InlineKeyboardMarkup, move *int) {
	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton := tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")
	// TODO: fix fields
	select {
	case updateCallback := <-cust.TranslateUpdate:
		chatID := updateCallback.Message.Chat.ID
		if updateCallback.Data == " " {
			fmt.Println("the occupied cell is pressed")

			replymsg := tgbotapi.NewMessage(chatID, "This cell is already occupied\nSelect another cell")

			msgData, err := bot.Send(replymsg)
			if err != nil {
				log.Println(err)
			}

			cust.OccupiedSells = append(cust.OccupiedSells, msgData.MessageID)

			return
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

		buttonMatrix.InlineKeyboard[row][column] = zeroButton
		cust.Players[updateCallback.From.ID].PlayingField[row][column] = (*move+1)%2 + 1
		*move++

		replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

		_, err = bot.Send(replymsg)
		if err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, updateCallback.From.ID, *move) {
			cust.Players[updateCallback.From.ID].RunGame = false
			cust.Players[updateCallback.From.ID].PlayingField = cust.NilPlayField
			return
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			if _, err := bot.Send(drawMessage); err != nil {
				log.Println(err)
			}

			cust.Players[updateCallback.From.ID].RunGame = false
			cust.Players[updateCallback.From.ID].PlayingField = cust.NilPlayField

			return
		}

		situation := Situation{PlayField: cust.NilPlayField}
		motion, _ := situation.Analyze((*move+1)%2+1, *move)

		buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
		cust.Players[updateCallback.From.ID].PlayingField[row][column] = (*move+1)%2 + 1
		*move++

		replymsg = tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

		time.Sleep(time.Millisecond * 300)

		_, err = bot.Send(replymsg)
		if err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, updateCallback.From.ID, *move) {
			cust.Players[updateCallback.From.ID].RunGame = false
			cust.Players[updateCallback.From.ID].PlayingField = cust.NilPlayField
			return
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			_, err = bot.Send(drawMessage)
			if err != nil {
				log.Println(err)
			}

			cust.Players[updateCallback.From.ID].RunGame = false
			cust.Players[updateCallback.From.ID].PlayingField = cust.NilPlayField
			fmt.Println("Draw")

			return
		}

	case Message := <-cust.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(int64(Message.From.ID), "Game stopped")

		if _, err := bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		cust.Players[Message.From.ID].RunGame = false
		cust.Players[Message.From.ID].PlayingField = cust.NilPlayField

		return
	}
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
