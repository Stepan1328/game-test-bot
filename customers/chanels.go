package customers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var (
	TranslateUpdate = make(chan tgbotapi.CallbackQuery)
	StopChannel     = make(chan tgbotapi.Message)
	Players         = make(map[int]*UsersStatistic)
	OccupiedSells   = make([]int, 0)
)
