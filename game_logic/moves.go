package gamelogic

import (
	"github.com/Stepan1328/game-test-bot/clients"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

// Tttgame is a function for starting a game with a bot
func Tttgame(update *tgbotapi.Update) {
	playerID := update.Message.From.ID
	msg := tgbotapi.NewMessage(clients.Players[playerID].ChatID, clients.Players[playerID].Location.Dictionary["main"])
	msg.ReplyMarkup = clients.Players[playerID].FieldMarkup

	msgData, err := clients.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	clients.Players[playerID].MsgID = msgData.MessageID
	clients.SaveBase()
}

// FirstMove is a function that makes the first move of the bot if necessary
func FirstMove(update *tgbotapi.Update) {
	var playerID int
	if update.Message != nil {
		playerID = update.Message.From.ID
		if !clients.Players[playerID].FirstMove && clients.Players[playerID].Field.Move == 1 {
			clients.Players[playerID].BotMove()
		}
	} else {
		playerID = update.CallbackQuery.From.ID
		if !clients.Players[playerID].FirstMove && clients.Players[playerID].Field.Move == 1 {
			clients.Players[playerID].BotMove()
		}
	}
}

// Motion is a function that processes callback responses
func Motion() {
	select {
	case updateCallback := <-clients.TranslateUpdate:
		if clients.Players[updateCallback.From.ID].CheckMsg(updateCallback.Message.MessageID) {
			makeDoubleMove(updateCallback)
			return
		}

		TemporaryMessage(updateCallback.From.ID, "irrelevant_field")
	case Message := <-clients.StopChannel:
		stopGameMessage := tgbotapi.NewMessage(clients.Players[Message.From.ID].ChatID, "Game stopped")

		if _, err := clients.Bot.Send(stopGameMessage); err != nil {
			log.Println(err)
		}

		clients.Players[Message.From.ID].ClearField()
		clients.SaveBase()

		clients.CheckInvitationStack(Message.From.ID)
	}
}

func makeDoubleMove(updateCallback tgbotapi.CallbackQuery) {
	playerID := updateCallback.From.ID
	if updateCallback.Data == " " {
		TemporaryMessage(playerID, "occupied_cell")
		clients.SaveBase()
		return
	}

	DeleteMessage(playerID)

	if clients.Players[playerID].HumanMove(updateCallback.Data) {
		return
	}

	clients.Players[playerID].BotMove()
}

// DeleteMessage is a function that deletes the message whose IDs are written to the array
func DeleteMessage(playerID int) {
	for len(clients.Players[playerID].OccupiedSells) > 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(clients.Players[playerID].ChatID, clients.Players[playerID].OccupiedSells[0])

		if _, err := clients.Bot.Send(deleteMsg); err != nil {
			log.Println(err)
		}

		clients.Players[playerID].OccupiedSells = clients.Players[playerID].OccupiedSells[1:]
	}
	clients.SaveBase()
}

// TemporaryMessage is a function for sending a message
// that will be deleted in the future using the DeleteMessage() function
func TemporaryMessage(playerID int, text string) {
	replymsg := tgbotapi.NewMessage(clients.Players[playerID].ChatID, clients.Players[playerID].Location.Dictionary[text])

	msgData, err := clients.Bot.Send(replymsg)
	if err != nil {
		log.Println(err)
	}

	clients.Players[playerID].OccupiedSells = append(clients.Players[playerID].OccupiedSells, msgData.MessageID)
	clients.SaveBase()
}

// SimpleMsg is a function for sending a message
func SimpleMsg(playerID int, text string) {
	msg := tgbotapi.NewMessage(clients.Players[playerID].ChatID, clients.Players[playerID].Location.Dictionary[text])

	if _, err := clients.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
