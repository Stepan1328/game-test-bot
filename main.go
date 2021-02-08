package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
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
		if update.CallbackQuery != nil {
			translateUpdate <- *update.CallbackQuery
			continue
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		fmt.Println(update.Message.Text)

		if update.Message.Command() == "tttgame" && !gameIsRunning {
			buttonMatrix := tttgame(update, *bot)
			go listenCallbackQuery(update, *bot, translateUpdate, stopChannel, buttonMatrix)
			gameIsRunning = true

			continue
		}

		if update.Message.Command() == "stopgame" {
			if gameIsRunning {
				stopChannel <- "stopgame"
				gameIsRunning = false
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

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command\nTry to write /tttgame to play tic tac toe")
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func tttgame(update tgbotapi.Update, bot tgbotapi.BotAPI) tgbotapi.InlineKeyboardMarkup {
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

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	return buttonMatrix
}

func listenCallbackQuery(update tgbotapi.Update, bot tgbotapi.BotAPI, translateUpdate chan tgbotapi.CallbackQuery, stopChannel chan string, buttonMatrix tgbotapi.InlineKeyboardMarkup) {
	var messageIDsOfOccupiedSells []int
	move := 1
	chatID := update.Message.Chat.ID

	playingField := [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton := tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")

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

				if move%2 == 1 {
					buttonMatrix.InlineKeyboard[row][column] = crossButton
					playingField[row][column] = 1
					move++
				} else {
					buttonMatrix.InlineKeyboard[row][column] = zeroButton
					playingField[row][column] = 2
					move++
				}

				replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, updateCallback.Message.MessageID, buttonMatrix)

				_, err = bot.Send(replymsg)
				if err != nil {
					log.Println(err)
				}

				if move > 9 {
					for i := 0; i < 3; i++ {
						fmt.Println(playingField[i])
					}

					endOfGameMsg := tgbotapi.NewMessage(chatID, "Draw ⚔️\nLucky next time\nLet's play one more timed")

					_, err = bot.Send(endOfGameMsg)
					if err != nil {
						log.Println(err)
					}
					return
				}
			case <-stopChannel:
				stopGameMessage := tgbotapi.NewMessage(chatID, "Game stopped")

				_, err := bot.Send(stopGameMessage)
				if err != nil {
					log.Println(err)
				}

				return
			}
		}
	}()
}
