package cust

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UsersStatistic struct {
	PlayerID      int //TODO: mute battle invite
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
	Mutex        bool
	PlayingField [3][3]int
	Move         int
}

func (user *UsersStatistic) CheckMsg(MsgID int) bool {
	if user.MsgID == MsgID {
		return true
	}

	return false
}

func (user *UsersStatistic) ClearField() {
	user.RunGame = false
	user.Field.PlayingField = [3][3]int{}
	user.Field.Move = 1
}

type BattleStatistic struct {
	FirstMove   bool
	Player1     *Player
	Player2     *Player
	Field       *Field
	FieldMarkup tgbotapi.InlineKeyboardMarkup
}

type Player struct {
	UserName string
	MsgID    int
	Queue    bool
	Score    int
}

func (user *BattleStatistic) ClearBattle() {
	user.Field.PlayingField = [3][3]int{}
	user.Field.Move = 1
}
