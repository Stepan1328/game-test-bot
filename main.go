package main

func main() {
	bot, updates := StartBot()

	ActionsWithUpdates(updates, bot)
}
