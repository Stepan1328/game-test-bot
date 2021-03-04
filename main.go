package main

import start "github.com/Stepan1328/game-test-bot/start_bot"

func main() {
	bot, updates := start.StartBot()

	start.ActionsWithUpdates(updates, bot)
}
