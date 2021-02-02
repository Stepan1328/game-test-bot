package main

import (
	"GameTestbot/YouTubeBot"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// entry point to the program
func main() {
	botToken := "1608392984:AAFZK4Rd-2Msm5RLJsLTCHDp5Bho6lGzujQ"
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi + botToken
	offset := 0

	err := startMessage(botUrl, offset)
	if err != nil {
		log.Println("Something went wrong: ", err.Error())
	}

	for {
		updates, err := getUpdates(botUrl, offset)
		if err != nil {
			log.Println("Something went wrong: ", err.Error())
		}
		for _, update := range updates {
			err = respond(botUrl, update)
			offset = update.UpdateId + 1
		}
		fmt.Println(updates)
	}
}

// request for updates
func getUpdates(botUrl string, offset int) ([]Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Result, nil
}

// response to updates
func respond(botUrl string, update Update) error {
	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	videoUrl, err := YouTubeBot.GetLastVideo(update.Message.Text)
	if err != nil {
		return err
	}

	botMessage.Text = videoUrl

	buf, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}

	_, err = http.Post(botUrl + "/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	return nil
}

func startMessage(botUrl string, offset int) error {
	updates, err := getUpdates(botUrl, offset)
	if err != nil {
		log.Println("Something went wrong: ", err.Error())
	}

	var botMessage BotMessage

	for _, update := range updates {
		botMessage.ChatId = update.Message.Chat.ChatId
		botMessage.Text = "Привет! \nЯ простой бот, который может повторять твои сообщения. \nПросто напиши мне сообщение, и я пришлю тебе его в ответ."
		//offset = update.UpdateId + 1
	}

	buf, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}

	_, err = http.Post(botUrl + "/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}