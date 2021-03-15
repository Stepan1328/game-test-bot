package customers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type UsersStatistic struct {
	RunGame       bool
	FirstMove     bool
	ChatID        int64
	MsgID         int
	Location      *Localization
	Field         *Field
	FieldMarkup   tgbotapi.InlineKeyboardMarkup
	OccupiedSells []int
}

type Localization struct {
	Language   string
	Dictionary map[string]string
}

type Field struct {
	PlayingField [3][3]int
	Move         int
}

func (u *UsersStatistic) CheckMsg(MsgID int) bool {
	if u.MsgID == MsgID {
		return true
	}

	return false
}

func (u *UsersStatistic) ClearField() {
	u.RunGame = false
	u.Field.PlayingField = [3][3]int{}
	u.Field.Move = 1
}

type BattleStatistic struct {
	RunGame     bool
	FirstMove   bool
	Player1     *Player
	Player2     *Player
	Field       *Field
	FieldMarkup tgbotapi.InlineKeyboardMarkup
}

type Player struct {
	UserName string
	ChatID   int64
	MsgID    int
	Queue    bool
	Location *Localization
}

func (u *BattleStatistic) ClearBattle() {
	u.RunGame = false
	u.Field.PlayingField = [3][3]int{}
	u.Field.Move = 1
}
