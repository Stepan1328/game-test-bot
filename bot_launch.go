package main

import (
	"fmt"
	"github.com/Stepan1328/game-test-bot/game_logic"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

var (
	translateUpdate           = make(chan tgbotapi.CallbackQuery)
	stopChannel               = make(chan tgbotapi.Message)
	gameIsRunningMap          = make(map[*tgbotapi.User]bool)
	playingFieldMap           = make(map[*tgbotapi.User][][]int)
	messageIDsOfOccupiedSells = make([]int, 0)
)

func ActionsWithUpdates(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
	for update := range updates {
		CheckUpdate(update, bot)
	}
}

func CheckUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if CheckPlayer(update) {
		return
	}

	Message := update.Message
	runGame := gameIsRunningMap[Message.From]
	fmt.Println(Message.Text)

	if Message.Command() != "" && !runGame {
		RecognitionCommand(update, bot)
		return
	}

	if runGame {
		UnfinishedGameMsg(Message, bot)
		return
	}

	if Message.Command() == "stopgame" {
		fmt.Println("stopgame")
		StopGame(Message, bot)
		return
	}
}

func CheckPlayer(update tgbotapi.Update) bool {
	if update.Message != nil {
		if _, inMap := gameIsRunningMap[update.Message.From]; !inMap {
			gameIsRunningMap[update.Message.From] = false
		}
	}

	if update.CallbackQuery != nil && gameIsRunningMap[update.CallbackQuery.From] {
		translateUpdate <- *update.CallbackQuery
		return true
	}

	return false
}

func GameLaunch(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

	buttonMatrix, msgID := tttgame(update, *bot)
	go listenCallbackQuery(update, *bot, buttonMatrix, msgID)

	gameIsRunningMap[update.Message.From] = true
}

func StopGame(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	if gameIsRunningMap[Message.From] {
		stopChannel <- *Message
	} else {
		msg := tgbotapi.NewMessage(Message.Chat.ID,
			"No game running\nTry to write /tttgame to play tic tac toe")

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func StartMsg(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(Message.Chat.ID,
		"Hi, if you want to play tic tac toe write /tttgame")

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func UnfinishedGameMsg(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(Message.Chat.ID,
		"Please finish playing the game or finish it by writing \n/stopgame")

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func RecognitionCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch update.Message.Command() {
	case "tttgame":
		GameLaunch(update, bot)
	case "start":
		StartMsg(update.Message, bot)
	}
}

func tttgame(update tgbotapi.Update, bot tgbotapi.BotAPI) (tgbotapi.InlineKeyboardMarkup, int) {
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

func listenCallbackQuery(update tgbotapi.Update, bot tgbotapi.BotAPI, buttonMatrix tgbotapi.InlineKeyboardMarkup, msgID int) {
	move := 1

	playingFieldMap[update.Message.From] = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")

	situation := game_logic.Situation{PlayField: playingFieldMap[update.Message.From]}
	motion, _ := situation.Analyze((move+1)%2+1, move)

	buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
	playingFieldMap[update.Message.From][motion.Y][motion.X] = (move+1)%2 + 1
	move++

	replymsg := tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID, msgID, buttonMatrix)

	time.Sleep(time.Millisecond * 200)

	if _, err := bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	go func() {
		for {
			MakeDoubleMove(bot, buttonMatrix, playingFieldMap[update.Message.From], &move)
		}
	}()
}

func MakeDoubleMove(bot tgbotapi.BotAPI, buttonMatrix tgbotapi.InlineKeyboardMarkup, playingField [][]int, move *int) {
	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton := tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")

	select {
	case updateCallback := <-translateUpdate:
		chatID := updateCallback.Message.Chat.ID
		if updateCallback.Data == " " {
			fmt.Println("the occupied cell is pressed")

			replymsg := tgbotapi.NewMessage(chatID, "This cell is already occupied\nSelect another cell")

			msgData, err := bot.Send(replymsg)
			if err != nil {
				log.Println(err)
			}

			messageIDsOfOccupiedSells = append(messageIDsOfOccupiedSells, msgData.MessageID)

			return
		}

		for len(messageIDsOfOccupiedSells) > 0 {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageIDsOfOccupiedSells[0])

			if _, err := bot.Send(deleteMsg); err != nil {
				log.Println(err)
			}

			messageIDsOfOccupiedSells = messageIDsOfOccupiedSells[1:]
			fmt.Println("message deleted")
		}

		numberOfCell, err := strconv.Atoi(updateCallback.Data)
		if err != nil {
			log.Println(err)
		}

		column := (numberOfCell - 1) % 3
		row := (numberOfCell - 1) / 3

		buttonMatrix.InlineKeyboard[row][column] = zeroButton
		playingField[row][column] = (*move+1)%2 + 1
		*move++

		replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

		_, err = bot.Send(replymsg)
		if err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, playingField, *move) {
			gameIsRunningMap[updateCallback.From] = false
			return
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			if _, err := bot.Send(drawMessage); err != nil {
				log.Println(err)
			}

			gameIsRunningMap[updateCallback.From] = false

			return
		}

		situation := game_logic.Situation{PlayField: playingField}
		motion, _ := situation.Analyze((*move+1)%2+1, *move)

		buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
		playingField[motion.Y][motion.X] = (*move+1)%2 + 1
		*move++

		replymsg = tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

		time.Sleep(time.Millisecond * 300)

		_, err = bot.Send(replymsg)
		if err != nil {
			log.Println(err)
		}

		if sendWinMsg(bot, chatID, playingField, *move) {
			gameIsRunningMap[updateCallback.From] = false
			return
		}

		if *move > 9 {
			drawMessage := tgbotapi.NewMessage(chatID,
				"Draw ⚔️\nLucky next time\nLet's play one more timed, write /tttgame")

			_, err = bot.Send(drawMessage)
			if err != nil {
				log.Println(err)
			}

			gameIsRunningMap[updateCallback.From] = false
			fmt.Println("Draw")

			return
		}

	case Message := <-stopChannel:
		stopGameMessage := tgbotapi.NewMessage(int64(Message.From.ID), "Game stopped")

		if _, err := bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		gameIsRunningMap[Message.From] = false

		return
	}
}

func sendWinMsg(bot tgbotapi.BotAPI, chatID int64, playingField [][]int, move int) bool {
	win, _ := game_logic.CheckingWinner(playingField, move-1)

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
