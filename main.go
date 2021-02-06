package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

const botToken = "1608392984:AAFZK4Rd-2Msm5RLJsLTCHDp5Bho6lGzujQ"

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

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		fmt.Println(update.Message.Text)

		if update.Message.Command() == "tttgame" {
			move := 1
			chatID := update.Message.Chat.ID

			msg := tgbotapi.NewMessage(chatID, "This is a tic-tac-toe field\nTo put a cross or a zero, just click on the button and its state will change\nGood luck")

			//playingField := [][]string {{"0", "0", "0"},{"0", "0", "0"},{"0", "0", "0"}}

			emptyButton := tgbotapi.NewInlineKeyboardButtonData(" ", " ")
			crossButton := tgbotapi.NewInlineKeyboardButtonData("❌", " ")
			zeroButton := tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")

			cellNumber := 1

			buttonRow1 := tgbotapi.NewInlineKeyboardRow()
			buttonRow2 := tgbotapi.NewInlineKeyboardRow()
			buttonRow3 := tgbotapi.NewInlineKeyboardRow()

			for i := 0; i < 3; i++ {
				*emptyButton.CallbackData = strconv.Itoa(cellNumber)
				buttonRow1 = append(buttonRow1, emptyButton)
				cellNumber++
			}

			for i := 0; i < 3; i++ {
				*emptyButton.CallbackData = strconv.Itoa(cellNumber)
				buttonRow2 = append(buttonRow1, emptyButton)
				cellNumber++
			}

			for i := 0; i < 3; i++ {
				*emptyButton.CallbackData = strconv.Itoa(cellNumber)
				buttonRow3 = append(buttonRow1, emptyButton)
				cellNumber++
			}

			buttonMatrix := tgbotapi.NewInlineKeyboardMarkup(buttonRow1, buttonRow2, buttonRow3)

			msg.ReplyMarkup = buttonMatrix

			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

			go func() {
				for {
					if update.CallbackQuery != nil {
						chatID = update.CallbackQuery.Message.Chat.ID
						if move%2 == 1 {
							buttonRow1 = tgbotapi.NewInlineKeyboardRow(crossButton, emptyButton, emptyButton)
							move++
						} else {
							buttonRow1 = tgbotapi.NewInlineKeyboardRow(zeroButton, emptyButton, emptyButton)
							move++
						}

						changeMatrix := tgbotapi.NewInlineKeyboardMarkup(buttonRow1, buttonRow2, buttonRow3)

						replymsg := tgbotapi.NewEditMessageReplyMarkup(chatID, update.CallbackQuery.Message.MessageID, changeMatrix)

						_, err = bot.Send(replymsg)
						if err != nil {
							log.Println(err)
						}

						fmt.Println(update.CallbackQuery.Data)

						update.CallbackQuery = nil
					}

					if move > 9 {
						endOfGameMsg := tgbotapi.NewMessage(chatID, "You are the winner 🎉")

						_, err = bot.Send(endOfGameMsg)
						if err != nil {
							log.Println(err)
						}
						return
					}
				}
			}()
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command\nTry to write /tttgame to play tic tac toe")
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
