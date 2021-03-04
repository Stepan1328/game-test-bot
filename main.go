package main

import (
	"bufio"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

func main() {
	bot, updates := startBot()

	ActionsWithUpdates(updates, bot)
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
	var botToken string

	file, err := os.Open("./botToken.txt")
	if err != nil {
		fmt.Println(err)
	}

	s := bufio.NewScanner(file)
	for s.Scan() {
		botToken = s.Text()
	}

	return botToken
}
