package main

import (
	"fmt"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const UNKNOWN_COMMAND string = "Error: Unknown Command!"

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("CollabLog_botKEY"))

	if err != nil {
		fmt.Println("Failed to initialize Bot API!")
		panic(err)
	}

	/* State */
	// appendMap := make(map[int64]string)
	state := NewState()

	/* Config and Setup Bot */
	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	/* Main processing loop */
	for update := range updates {
		if update.Message == nil {
			continue
		}

		/* Handle Command */
		if update.Message.IsCommand() {
			receivedCommand := update.Message.Command()

			switch receivedCommand {
			case "new": /* Handle make new Log */
				reply := state.MakeNewLog(update)
				/* Reply to user */
				sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)

			case "created": /* List all logs I created */
				reply := state.GetCreatedLogs(update)
				sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)

			case "edit": /* Edit (Overwrite) an existing log */
				sendReply(bot, state.EditLog(update), update.Message.Chat.ID, update.Message.MessageID)

			case "read": /* Read current state of existing log */
				sendReply(bot, state.ReadLog(update), update.Message.Chat.ID, update.Message.MessageID)

			case "delete": /* Delete an existing log */
				sendReply(bot, state.DeleteLog(update), update.Message.Chat.ID, update.Message.MessageID)

			default: /* Unrecognized command */
				sendReply(bot, UNKNOWN_COMMAND, update.Message.Chat.ID, update.Message.MessageID)
			}
		}

		if !update.Message.IsCommand() {
			// if _, ok := appendMap[update.Message.Chat.ID]; !ok {
			// 	appendMap[update.Message.Chat.ID] = ""
			// }

			// reply := "<b>Received</b> message from [@" + update.Message.From.UserName + "]:\n" + update.Message.Text
			// appendMap[update.Message.Chat.ID] += "\n" + update.Message.Text
			// reply += "\nMessages seen so far by [@" + update.Message.From.UserName + "]:\n" + appendMap[update.Message.Chat.ID]

			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			// msg.ParseMode = "HTML"
			// msg.ReplyToMessageID = update.Message.MessageID

			// if _, err := bot.Send(msg); err != nil {
			// 	panic(err)
			// }
		}

	}
}

func sendReply(bot *tgbotapi.BotAPI, reply string, chatID int64, messageID int) {
	msg := tgbotapi.NewMessage(chatID, reply)
	msg.ParseMode = "HTML"
	msg.ReplyToMessageID = messageID

	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d | %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}
