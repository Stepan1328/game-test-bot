package main

import (
	"github.com/Stepan1328/game-test-bot/clients"
	"github.com/Stepan1328/game-test-bot/game_logic"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

func main() {
	var updates tgbotapi.UpdatesChannel
	clients.UploadDatabase()

	startHandler()
	clients.Bot, updates = startBot()

	ActionsWithUpdates(updates)
}

func startBot() (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
	botToken := takeBotToken()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic("Failed to initialize bot: " + err.Error())
	}

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
			game_logic.Motion()
		}
	}()

	go func() {
		for {
			game_logic.BattleMotion()
		}
	}()
}
