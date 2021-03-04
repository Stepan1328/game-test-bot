package customers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type UsersStatistic struct {
	RunGame      bool
	PlayingField [3][3]int
	FieldMarkup  tgbotapi.InlineKeyboardMarkup
}

func (u *UsersStatistic) ClearField() {
	u.RunGame = false
	u.PlayingField = [3][3]int{}
}
