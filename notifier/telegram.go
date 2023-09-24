package notifier

import (
	"dhtc/config"
	"fmt"
	telegram "gopkg.in/telebot.v3"
	"time"
)

func SetupTelegramBot(config *config.Configuration) *telegram.Bot {
	var rVal *telegram.Bot = nil
	if config.TelegramToken != "" {
		pref := telegram.Settings{
			Token:  config.TelegramToken,
			Poller: &telegram.LongPoller{Timeout: 10 * time.Second},
		}

		var err error
		rVal, err = telegram.NewBot(pref)
		if err != nil {
			fmt.Println("Could not create telegram bot.")
		}
	}
	return rVal
}

func NotifyTelegram(config *config.Configuration, bot *telegram.Bot, message string) {
	if bot != nil {
		chat, err := bot.ChatByUsername(config.TelegramUsername)
		if err == nil {
			fmt.Println("Sending telegram notification.")
			_, err := bot.Send(chat, message)
			if err != nil {
				fmt.Printf("Could not send message '%s' to user '%s'.", message, config.TelegramUsername)
			}
		} else {
			fmt.Printf("Could not find chat by username '%s'\n", config.TelegramUsername)
		}
	}
}
