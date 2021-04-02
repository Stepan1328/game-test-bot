package clients

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// Bot is a structure used for sending messages to chats
	Bot               *tgbotapi.BotAPI

	// TranslateUpdate is a channel for sending click responses
	TranslateUpdate   = make(chan tgbotapi.CallbackQuery)
	
	// TranslateBattle is a channel for sending responses related to battles
	TranslateBattle   = make(chan tgbotapi.CallbackQuery)

	// StopChannel is a channel for sending stop commands
	StopChannel       = make(chan tgbotapi.Message)
	
	// StopBattleChannel is a channel for sending messages stopping the battle
	StopBattleChannel = make(chan tgbotapi.Message)
	
	// Players is a PlayerBase map which assigns all the user IDs to their structures
	Players           = make(map[int]*UsersStatistic)
	
	// Compatibility is a map which stores the link between the player's nickname
	// and his ID to find it in the player database
	Compatibility     = make(map[string]int)
	
	// Battles is a map which stores battles between players
	// the key is to combine the nicknames of two players
	Battles           = make(map[string]*BattleStatistic)
)
