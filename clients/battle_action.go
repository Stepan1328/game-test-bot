package clients

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HumanBattleMove is a structure that allows you to easily send
// and change the battle fields after a response from the keyboard
func (user *BattleStatistic) HumanBattleMove(data string) {
	if !user.RunGame {
		return
	}
	numberOfCell, err := strconv.Atoi(data)
	if err != nil {
		log.Println(err)
	}

	user.EditField(numberOfCell)
	user.ParseMarkUp()

	user.Player1.EditMarkUpMsg()
	user.Player2.EditMarkUpMsg()

	user.Player1.Queue, user.Player2.Queue =
		user.Player2.Queue, user.Player1.Queue

	if user.CheckSituation() {
		user.Player1.Queue = user.FirstMove
		user.Player2.Queue = !user.FirstMove
		return
	}

	user.editTitleOfMsg()
	SaveBase()
}

func (user *BattleStatistic) editTitleOfMsg() {
	if user.FirstMove {
		user.Player1.sendEditTitle(1)
		user.Player2.sendEditTitle(2)
		return
	}
	user.Player1.sendEditTitle(2)
	user.Player2.sendEditTitle(1)
}

func (player *Player) sendEditTitle(offset int) {
	var text string
	switch offset {
	case 1:
		if player.Queue {
			text = Players[player.PlayerID].Location.Dictionary["battle_cross_his_turn"]
		} else {
			text = Players[player.PlayerID].Location.Dictionary["battle_cross_no_his_turn"]
		}
	case 2:
		if player.Queue {
			text = Players[player.PlayerID].Location.Dictionary["battle_zero_his_turn"]
		} else {
			text = Players[player.PlayerID].Location.Dictionary["battle_zero_no_his_turn"]
		}
	}

	editMsg := tgbotapi.NewEditMessageText(Players[player.PlayerID].ChatID, player.MsgID, text)
	userName := Players[player.PlayerID].UserName
	editMsg.ReplyMarkup = &Battles[userName].FieldMarkup

	if _, err := Bot.Send(editMsg); err != nil {
		log.Println(err)
	}
}

// EditMarkUpMsg is method used to change a field in a message
func (player *Player) EditMarkUpMsg() {
	replyMsg := tgbotapi.NewEditMessageReplyMarkup(Players[player.PlayerID].ChatID,
		player.MsgID, Battles[player.UserName].FieldMarkup)

	if _, err := Bot.Send(replyMsg); err != nil {
		log.Println(err)
	}
}

// EditField is a method with which you can conveniently change the fields of the battle structure
func (user *BattleStatistic) EditField(numberOfCell int) {
	column := (numberOfCell - 1) % 3
	row := (numberOfCell - 1) / 3
	user.Field.PlayingField[row][column] = (user.Field.Move+1)%2 + 1
	user.Field.Move++
}

// ParseMarkUp is a convenient method for parsing a field-keyboard from its field
func (user *BattleStatistic) ParseMarkUp() {
	var masOfButton [9]tgbotapi.InlineKeyboardButton
	var masOfRow [3][]tgbotapi.InlineKeyboardButton

	for i := 1; i <= 9; i++ {
		switch user.Field.PlayingField[(i-1)/3][(i-1)%3] {
		case 0:
			button := tgbotapi.NewInlineKeyboardButtonData(" ", strconv.Itoa(i))
			masOfButton[i-1] = button
		case 1:
			masOfButton[i-1] = crossButton
		case 2:
			masOfButton[i-1] = zeroButton
		}
	}

	for i := 0; i < 3; i++ {
		buttonRow := tgbotapi.NewInlineKeyboardRow(masOfButton[i*3], masOfButton[i*3+1], masOfButton[i*3+2])
		masOfRow[i] = buttonRow
	}

	user.FieldMarkup = tgbotapi.NewInlineKeyboardMarkup(masOfRow[0], masOfRow[1], masOfRow[2])
}

// CheckSituation is a method who checks the situation on the field
func (user *BattleStatistic) CheckSituation() bool {
	if user.sendWinMsg() {
		user.ClearField()
		if user.FirstMove {
			user.FirstMove = false
		} else {
			user.FirstMove = true
		}
		return true
	}

	if user.Field.Move > 9 {
		user.sendDrawMsg()
		user.ClearField()
		if user.FirstMove {
			user.FirstMove = false
		} else {
			user.FirstMove = true
		}
		return true
	}
	return false
}

// CheckQueue checks the order of moves and access to the field change
func (user *BattleStatistic) CheckQueue(userName string) bool {
	if user.Player1.UserName == userName {
		return user.Player1.Queue
	}
	return user.Player2.Queue
}

func (user *BattleStatistic) sendDrawMsg() {
	user.Player1.sendBattleDrawMsg()
	user.Player2.sendBattleDrawMsg()

	user.ClearField()
}

func (player *Player) sendBattleDrawMsg() {
	drawMessage := tgbotapi.NewMessage(Players[player.PlayerID].ChatID, Players[player.PlayerID].Location.Dictionary["battle_draw"])

	if _, err := Bot.Send(drawMessage); err != nil {
		log.Println(err)
	}
}

func (user *BattleStatistic) sendWinMsg() bool {
	win, _ := CheckingWinner(user.Field.PlayingField, user.Field.Move-1)

	if !win {
		return win
	}

	move := user.Field.Move
	user.Player1.sendBattleWinMsg(move)
	user.Player2.sendBattleWinMsg(move)
	user.increaseScore()
	fmt.Println("found the winner")

	return win
}

func (player *Player) sendBattleWinMsg(move int) {
	winMessage := tgbotapi.NewMessage(Players[player.PlayerID].ChatID, "")

	if move%2 == 0 {
		winMessage.Text = Players[player.PlayerID].Location.Dictionary["battle_win_cross"]
	} else {
		winMessage.Text = Players[player.PlayerID].Location.Dictionary["battle_win_zero"]
	}

	if _, err := Bot.Send(winMessage); err != nil {
		log.Println(err)
	}
}

func (user *BattleStatistic) increaseScore() {
	if user.FirstMove {
		if (user.Field.Move-1)%2 == 1 {
			user.Player1.Score++
		} else {
			user.Player2.Score++
		}
	} else {
		if (user.Field.Move-1)%2 == 1 {
			user.Player2.Score++
		} else {
			user.Player1.Score++
		}
	}
}
