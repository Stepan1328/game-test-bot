package clients

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UsersStatistic is the main structure containing all the settings and statistics of a particular user
type UsersStatistic struct {
	UserName        string
	SettingID       int
	BattleInvite    bool
	LastBattleID    int64
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

// Localization responsible for the language in which messages are sent
type Localization struct {
	Language   string
	Dictionary map[string]string
}

// Field stores the state of the field
type Field struct {
	PlayingField [3][3]int
	Move         int
}

// CheckMsg is a method that checks the correctness of the message id
func (user *UsersStatistic) CheckMsg(MsgID int) bool {
	if user.MsgID == MsgID {
		return true
	}

	return false
}

// ClearField completely clears the field after the game
func (user *UsersStatistic) ClearField() {
	user.RunGame = false
	user.Field.PlayingField = [3][3]int{}
	user.Field.Move = 1

	user.ParseMarkUp()
}

// BattleStatistic the main structure containing all the information about
// a couple of users who played the battle
type BattleStatistic struct {
	RunGame     bool
	FirstMove   bool
	Player1     *Player
	Player2     *Player
	Field       *Field
	FieldMarkup tgbotapi.InlineKeyboardMarkup
}

// Player is a structure containing information about one of the players in the battle
type Player struct {
	UserName string
	PlayerID int64
	MsgID    int
	Queue    bool
	Score    int
}

// CheckMsg is a method that checks the correctness of the message id
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

// ClearField completely clears the field after the game
func (user *BattleStatistic) ClearField() {
	user.RunGame = false
	user.Field.PlayingField = [3][3]int{}
	user.Field.Move = 1

	user.ParseMarkUp()
}
