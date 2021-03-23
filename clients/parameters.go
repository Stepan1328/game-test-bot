package clients

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UsersStatistic struct {
	UserName        string
	BattleInvite    bool
	RunGame         bool
	FirstMove       bool
	ChatID          int64
	MsgID           int
	Location        *Localization
	Field           *Field
	FieldMarkup     tgbotapi.InlineKeyboardMarkup
	OccupiedSells   []int
	InvitationStack []tgbotapi.MessageConfig
}

type Localization struct {
	Language   string
	Dictionary map[string]string
}

type Field struct {
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
	user.ParseMarkUp()
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
	PlayerId int
	MsgID    int
	Queue    bool
	Score    int
}

func (user *BattleStatistic) CheckMsg(MsgID int, userName string) bool {
	var trueMsgID int
	if Battles[userName].Player1.UserName == userName {
		trueMsgID = Battles[userName].Player1.MsgID
	} else {
		trueMsgID = Battles[userName].Player2.MsgID
	}

	if trueMsgID == MsgID {
		return true
	}
	return false
}

func (user *BattleStatistic) ClearField() {
	user.RunGame = false
	user.Field.PlayingField = [3][3]int{}
	user.Field.Move = 1
}
