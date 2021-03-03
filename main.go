package main

import "github.com/Stepan1328/game-test-bot/start_bot"

func main() {
	bot, updates := start_bot.StartBot()

	start_bot.ActionsWithUpdates(updates, bot)
}
