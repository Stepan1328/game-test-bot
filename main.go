package main

import (
	"encoding/json"
	cust "github.com/Stepan1328/game-test-bot/customers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

func main() {
	parseMap()
	bot, updates := startBot()

	ActionsWithUpdates(&updates, bot)
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

func parseMap() {
	bytes, _ := os.ReadFile("./assets/en.json")

	_ = json.Unmarshal(bytes, &cust.LangMap)
}
