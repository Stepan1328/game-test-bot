package customers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var (
	TranslateUpdate = make(chan tgbotapi.CallbackQuery)
	TranslateBattle = make(chan tgbotapi.CallbackQuery)
	StopChannel     = make(chan tgbotapi.Message)
	Players         = make(map[int]*UsersStatistic)
	Battles         = make(map[int]*BattleStatistic)
	OccupiedSells   = make([]int, 0)
)
