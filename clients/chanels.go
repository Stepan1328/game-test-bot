package clients

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	Bot               *tgbotapi.BotAPI
	TranslateUpdate   = make(chan tgbotapi.CallbackQuery)
	TranslateBattle   = make(chan tgbotapi.CallbackQuery)
	StopChannel       = make(chan tgbotapi.Message)
	StopBattleChannel = make(chan tgbotapi.Message)
	Players           = make(map[int]*UsersStatistic)
	Compatibility     = make(map[string]int)
	Battles           = make(map[string]*BattleStatistic)
)
