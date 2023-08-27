package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	botToken   string = os.Getenv("BOT_TOKEN")
	guidelines string = os.Getenv("GUIDELINES")
	chatID     int64
)

func init() {
	i, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	chatID = i
}

func main() {
	bot := bot()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-stop:
	default:
		for update := range updates {
			if update.Message == nil {
				continue
			}
			handleUpdate(bot, update.Message)
		}
	}

	log.Println("Graceful shutdown")
}

func bot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = true

	return bot
}

func handleUpdate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	var response []string

	if message.IsCommand() {
		switch cmd := message.Command(); cmd {
		case "start":
			response = []string{
				"Hey there!\n\nTo submit, just send me an uncompressed .jpg or .png file.",
				fmt.Sprintf("Before submitting, please see the guidelines.\n\n%s", guidelines),
			}
		case "guidelines":
			response = []string{guidelines}
		default:
			response = []string{"Please enter a valid command. Type '/' into your chat box or " +
				"press 'Menu' button to see the list of available commands."}
		}
	} else if message.Document != nil {
		response = []string{"Thank you! Your application will be reviewed soon."}
		doc := tgbotapi.NewDocumentShare(chatID, message.Document.FileID)
		doc.Caption = fmt.Sprintf("From %s (%d)", message.Chat.UserName, message.Chat.ID)
		bot.Send(doc)
	} else {
		response = []string{"Please send a picture as a document or enter a valid command."}
	}

	for _, s := range response {
		msg := tgbotapi.NewMessage(message.Chat.ID, s)
		bot.Send(msg)
	}
}
