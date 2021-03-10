package customers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type UsersStatistic struct {
	RunGame     bool
	FirstMove   bool
	ChatID      int64
	Location    *Localization
	Field       *Field
	FieldMarkup tgbotapi.InlineKeyboardMarkup
}

type Localization struct {
	Language   string
	Dictionary map[string]string
}

type Field struct {
	PlayingField [3][3]int
	Move         int
}

func (u *UsersStatistic) ClearField() {
	u.RunGame = false
	u.Field.PlayingField = [3][3]int{}
	u.Field.Move = 1
}
