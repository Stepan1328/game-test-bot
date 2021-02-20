package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

const botToken = "1608392984:AAFKp0yFvDn5ZAgpKBBHX5soc7CIqXg9Z0I"

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic("Failed to initialize bot: " + err.Error())
	}

	translateUpdate := make(chan tgbotapi.CallbackQuery)
	stopChannel := make(chan string)
	gameIsRunning := false

	log.Println("The bot is running")

	for update := range updates {
		if update.CallbackQuery != nil && gameIsRunning {
			translateUpdate <- *update.CallbackQuery
			continue
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		fmt.Println(update.Message.Text)

		if update.Message.Command() == "tttgame" && !gameIsRunning {
			fmt.Println(update.Message.From.FirstName, update.Message.From.LastName)

			buttonMatrix, msgID := tttgame(update, *bot)
			go listenCallbackQuery(update, *bot, translateUpdate, stopChannel, buttonMatrix, msgID, &gameIsRunning)

			gameIsRunning = true

			continue
		}

		if update.Message.Command() == "stopgame" {
			if gameIsRunning {
				stopChannel <- "stopgame"
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "No game running\nTry to write /tttgame to play tic tac toe")
				_, err = bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}

			continue
		}

		if gameIsRunning {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please finish playing the game or finish it by writing \n/stopgame")
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

			continue
		}

		if update.Message.Command() == "start" && !gameIsRunning {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, if you want to play tic tac toe write /tttgame")
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

			continue
		}
	}
}

func tttgame(update tgbotapi.Update, bot tgbotapi.BotAPI) (tgbotapi.InlineKeyboardMarkup, int) {
	chatID := update.Message.Chat.ID

	msg := tgbotapi.NewMessage(chatID, "This is a tic-tac-toe field\nTo put a cross or a zero, just click on the button and its state will change\nGood luck")

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

func listenCallbackQuery(update tgbotapi.Update, bot tgbotapi.BotAPI, translateUpdate chan tgbotapi.CallbackQuery, stopChannel chan string, buttonMatrix tgbotapi.InlineKeyboardMarkup, msgID int, gameIsRunning *bool) {
	var messageIDsOfOccupiedSells []int
	move := 1
	chatID := update.Message.Chat.ID

	playingField := [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton := tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")

	situation := Situation{PlayField: playingField}
	motion, _ := situation.Analyze((move+1)%2+1, move)

	buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
	playingField[motion.Y][motion.X] = (move+1)%2 + 1
	move++

	replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, msgID, buttonMatrix)

	time.Sleep(time.Millisecond * 200)

	_, err := bot.Send(replymsg)
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			select {
			case updateCallback := <-translateUpdate:
				if updateCallback.Data == " " {
					fmt.Println("the occupied cell is pressed")

					replymsg := tgbotapi.NewMessage(chatID, "This cell is already occupied\nSelect another cell")

					msgData, err := bot.Send(replymsg)
					if err != nil {
						log.Println(err)
					}

					messageIDsOfOccupiedSells = append(messageIDsOfOccupiedSells, msgData.MessageID)

					continue
				}

				for len(messageIDsOfOccupiedSells) > 0 {
					deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageIDsOfOccupiedSells[0])

					_, err := bot.Send(deleteMsg)
					if err != nil {
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

				//if move % 2 == 1 {
				//	buttonMatrix.InlineKeyboard[row][column] = crossButton
				//	playingField[row][column] = 1
				//	move++
				//} else {
				//	buttonMatrix.InlineKeyboard[row][column] = zeroButton
				//	playingField[row][column] = 2
				//	move++
				//}
				//if move % 2 == 1 {
				//	buttonMatrix.InlineKeyboard[row][column] = crossButton
				//	playingField[row][column] = (move + 1) % 2 + 1
				//	move++
				//} else {
				//	situation := Situation{PlayField: playingField}
				//	motion, _ := situation.Analyze((move + 1) % 2 + 1, move)
				//
				//	buttonMatrix.InlineKeyboard[motion.Y][motion.X] = zeroButton
				//	playingField[motion.Y][motion.X] = (move + 1) % 2 + 1
				//	move++
				//}

				buttonMatrix.InlineKeyboard[row][column] = zeroButton
				playingField[row][column] = (move+1)%2 + 1
				move++

				replymsg = tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

				_, err = bot.Send(replymsg)
				if err != nil {
					log.Println(err)
				}

				win := sendWinMsg(bot, chatID, playingField, move, gameIsRunning)
				if win {
					return
				}

				if move > 9 {
					drawMessage := tgbotapi.NewMessage(chatID, "Draw ⚔️\nLucky next time\nLet's play one more timed")

					_, err = bot.Send(drawMessage)
					if err != nil {
						log.Println(err)
					}

					*gameIsRunning = false

					return
				}

				situation = Situation{PlayField: playingField}
				motion, _ = situation.Analyze((move+1)%2+1, move)

				buttonMatrix.InlineKeyboard[motion.Y][motion.X] = crossButton
				playingField[motion.Y][motion.X] = (move+1)%2 + 1
				move++

				replymsg = tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

				time.Sleep(time.Millisecond * 300)

				_, err = bot.Send(replymsg)
				if err != nil {
					log.Println(err)
				}

				win = sendWinMsg(bot, chatID, playingField, move, gameIsRunning)
				if win {
					return
				}

				if move > 9 {
					drawMessage := tgbotapi.NewMessage(chatID, "Draw ⚔️\nLucky next time\nLet's play one more timed")

					_, err = bot.Send(drawMessage)
					if err != nil {
						log.Println(err)
					}

					*gameIsRunning = false
					fmt.Println("Draw")

					return
				}

			case <-stopChannel:
				stopGameMessage := tgbotapi.NewMessage(chatID, "Game stopped")

				_, err = bot.Send(stopGameMessage)
				if err != nil {
					log.Println(err)
				}

				*gameIsRunning = false

				return
			}
		}
	}()
}

func sendWinMsg(bot tgbotapi.BotAPI, chatID int64, playingField [][]int, move int, gameIsRunning *bool) bool {
	win, _ := checkingWinner(playingField, move-1)

	if win {
		winMessage := tgbotapi.NewMessage(chatID, "")

		if move%2 == 0 {
			winMessage.Text = "Congratulations to 🔥Player 1🔥 he is the winner 🎉\nPlayer 2 don't get upset and ask for a rematch by writing /tttgame"
		} else {
			winMessage.Text = "Congratulations to 🔥Player 2🔥 he is the winner 🎉\nPlayer 1 don't get upset and ask for a rematch by writing /tttgame"
		}

		_, err := bot.Send(winMessage)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("found the winner")
		*gameIsRunning = false
		return win
	}

	return win
}
