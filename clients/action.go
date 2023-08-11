package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	crossButton = tgbotapi.NewInlineKeyboardButtonData("❌", " ")
	zeroButton  = tgbotapi.NewInlineKeyboardButtonData("⭕️", " ")
)

// Action is a move structure row is Y, column is X.
type Action struct {
	Y int
	X int
}

// Analyze is bot intelect.
func (f *Field) Analyze(player int, motion int) (Action, int) {
	winMoves := make([]Action, 0)
	drawMoves := make([]Action, 0)
	losingMoves := make([]Action, 0)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if f.PlayingField[i][j] == 0 && (motion-1)%2+1 == player {
				f.PlayingField[i][j] = (motion-1)%2 + 1

				win, _ := CheckingWinner(f.PlayingField, motion)
				if win {
					winMoves = append(winMoves, Action{i, j})
				}

				f.PlayingField[i][j] = 0
			}
		}
	}

	if len(winMoves) > 0 {
		return winMoves[rand.Intn(len(winMoves))], 2
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if f.PlayingField[i][j] == 0 {
				f.PlayingField[i][j] = (motion-1)%2 + 1

				win, winPlayer := CheckingWinner(f.PlayingField, motion)
				if win && winPlayer == player {
					winMoves = append(winMoves, Action{i, j})
				} else if win {
					losingMoves = append(losingMoves, Action{i, j})
				} else if motion == 9 {
					drawMoves = append(drawMoves, Action{i, j})
				} else {
					move, result := f.Analyze(player, motion+1)

					switch result {
					case 2:
						winMoves = append(winMoves, move)
					case 1:
						drawMoves = append(drawMoves, move)
					case 0:
						losingMoves = append(losingMoves, move)
					}
				}

				f.PlayingField[i][j] = 0
			}
		}
	}

	if player == (motion+1)%2+1 {
		if len(winMoves) > 0 {
			return winMoves[rand.Intn(len(winMoves))], 2
		}

		if len(drawMoves) > 0 {
			return drawMoves[rand.Intn(len(drawMoves))], 1
		}

		return losingMoves[rand.Intn(len(losingMoves))], 0
	}

	if len(losingMoves) > 0 {
		return losingMoves[rand.Intn(len(losingMoves))], 2
	}

	if len(drawMoves) > 0 {
		return drawMoves[rand.Intn(len(drawMoves))], 1
	}

	return winMoves[rand.Intn(len(winMoves))], 0
}

// HumanMove is a method to make human move.
func (user *UsersStatistic) HumanMove(data string) bool {
	numberOfCell, err := strconv.Atoi(data)
	if err != nil {
		log.Println(err)
	}

	user.EditField(numberOfCell)
	user.ParseMarkUp()

	replymsg := tgbotapi.NewEditMessageReplyMarkup(user.ChatID, user.MsgID, user.FieldMarkup)

	if _, err = Bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	if user.CheckSituation() {
		SaveBase()
		CheckInvitationStack(Compatibility[user.UserName])
		return true
	}

	SaveBase()
	return false
}

// BotMove is a method to make bot move.
func (user *UsersStatistic) BotMove() bool {
	motion, _ := user.Field.Analyze((user.Field.Move+1)%2+1, user.Field.Move)
	user.Field.PlayingField[motion.Y][motion.X] = (user.Field.Move+1)%2 + 1
	user.Field.Move++
	user.ParseMarkUp()

	replymsg := tgbotapi.NewEditMessageReplyMarkup(user.ChatID, user.MsgID, user.FieldMarkup)

	if _, err := Bot.Send(replymsg); err != nil {
		log.Println(err)
	}

	if user.CheckSituation() {
		SaveBase()
		if user.UserName != "" {
			CheckInvitationStack(Compatibility[user.UserName])
		}
		return true
	}

	SaveBase()
	return false
}

// CheckInvitationStack send invitation msg to player if he has pending requests
func CheckInvitationStack(userID int64) {
	if len(Players[userID].InvitationStack) > 0 {
		sendInvitationMsg(userID)
		Players[userID].InvitationStack = Players[userID].InvitationStack[1:]
	}
}

func sendInvitationMsg(userID int64) {
	msg := Players[userID].InvitationStack[0]
	msgData, err := Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	Players[userID].OccupiedSells = append(Players[userID].OccupiedSells, msgData.MessageID)
	SaveBase()
}

// EditField is a convenient method for changing the state of a player's field
func (user *UsersStatistic) EditField(numberOfCell int) {
	column := (numberOfCell - 1) % 3
	row := (numberOfCell - 1) / 3
	user.Field.PlayingField[row][column] = (user.Field.Move+1)%2 + 1
	user.Field.Move++
}

// ParseMarkUp is a convenient method for parsing a field-keyboard from its field
func (user *UsersStatistic) ParseMarkUp() {
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
func (user *UsersStatistic) CheckSituation() bool {
	if user.sendWinMsg() {
		user.ClearField()
		SaveBase()
		return true
	}

	if user.Field.Move > 9 {
		user.sendDrawMsg()
		return true
	}

	return false
}

func (user *UsersStatistic) sendDrawMsg() {
	drawMessage := tgbotapi.NewMessage(user.ChatID, user.Location.Dictionary["draw"])

	if _, err := Bot.Send(drawMessage); err != nil {
		log.Println(err)
	}

	user.ClearField()
	SaveBase()
}

func (user *UsersStatistic) sendWinMsg() bool {
	win, _ := CheckingWinner(user.Field.PlayingField, user.Field.Move-1)

	if !win {
		return win
	}

	winMessage := tgbotapi.NewMessage(user.ChatID, "")

	if user.Field.Move%2 == 0 {
		winMessage.Text = user.Location.Dictionary["win_cross"]
	} else {
		winMessage.Text = user.Location.Dictionary["win_zero"]
	}

	if _, err := Bot.Send(winMessage); err != nil {
		log.Println(err)
	}

	fmt.Println("found the winner")

	return win
}

// ParseLangMap is a method that changes the user's language dictionary
func (user *UsersStatistic) ParseLangMap() {
	lang := user.Location.Language
	bytes, _ := os.ReadFile("./assets/" + lang + ".json")

	_ = json.Unmarshal(bytes, &user.Location.Dictionary)
}
