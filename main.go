package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Stepan1328/game-test-bot/clients"
	"github.com/Stepan1328/game-test-bot/game_logic"
)

func main() {
	var updates tgbotapi.UpdatesChannel
	clients.UploadDatabase()

	startHandler()
	clients.Bot, updates = startBot()

	actionsWithUpdates(updates)
}

func startBot() (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
	botToken := takeBotToken()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)

	log.Println("The bot is running")

	return bot, updates
}

func takeBotToken() string {
	content, _ := os.ReadFile("./botToken.txt")
	return string(content)
}

func startHandler() {
	go func() {
		for {
			gamelogic.Motion()
		}
	}()

	go func() {
		for {
			gamelogic.BattleMotion()
		}
	}()
}
