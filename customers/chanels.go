package cust

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	Bot             *tgbotapi.BotAPI
	TranslateUpdate = make(chan tgbotapi.CallbackQuery)
	TranslateBattle = make(chan tgbotapi.CallbackQuery)
	StopChannel     = make(chan tgbotapi.Message)
	Players         = make(map[string]*UsersStatistic)
	Battles         = make(map[string]*BattleStatistic)
)
