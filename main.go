package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/google/uuid"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("CollabLog_botKEY"))

	if err != nil {
		fmt.Println("Failed to initialize Bot API!")
		panic(err)
	}

	/* State */
	// appendMap := make(map[int64]string)
	logOwnerMap := make(map[int64]map[string]*Log)
	editedLogsMap := make(map[int64]map[string]*Log)
	allLogs := make(map[string]*Log)

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
				args := strings.TrimSpace(update.Message.CommandArguments())

				/* Todo: HANDLE DUPLICATE LOG NAMES */

				/* Name of Log cannot be empty */
				if args == "" {
					errorMsg := "Error: name of Log cannot be empty.\n\n<b>Sample Usage:</b>\n/new MyLogName"
					sendReply(bot, errorMsg, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}
				newLogToAdd := newLog(args, *update.Message.From)

				/* Store as creator */
				if _, ok := logOwnerMap[update.Message.From.ID]; !ok {
					logOwnerMap[update.Message.From.ID] = map[string]*Log{}
				}
				logOwnerMap[update.Message.From.ID][args] = newLogToAdd

				/* Store as editor */
				if _, ok := editedLogsMap[update.Message.From.ID]; !ok {
					editedLogsMap[update.Message.From.ID] = map[string]*Log{}
				}
				editedLogsMap[update.Message.From.ID][args] = newLogToAdd

				allLogs[args] = newLogToAdd

				reply := fmt.Sprintf("Success! Created new Log <b>[%s]</b>\n", args)
				/* Reply to user */
				sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)

			case "created": /* List all logs I created */
				reply := fmt.Sprintf("<b>Logs created by @%s</b>:\n", update.Message.From.UserName)
				userID := update.Message.From.ID

				if _, ok := logOwnerMap[userID]; !ok {
					reply += "<i>No Logs found</i>\n"
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				if len(logOwnerMap[userID]) == 0 {
					reply += "<i>No Logs found</i>\n"
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				for key := range logOwnerMap[userID] {
					reply += key + "\n"
				}
				sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)

			case "edit":
				tokens := strings.Split(strings.TrimSpace(update.Message.CommandArguments()), " ")

				/* Check usage */
				if len(tokens) < 2 {
					reply := "Error: Wrong usage.\n\n<b>Sample Usage:</b>\n/edit LogName NewText"
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				logTitle := tokens[0]
				logNewText := strings.Join(tokens[1:], " ")
				editor := update.Message.From

				/* Check if the Log exists */
				if _, ok := allLogs[logTitle]; !ok {
					reply := fmt.Sprintf("Error: No Log with name <b>%s</b> exists!", logTitle)
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				/* Update Log */
				var logToEdit *Log
				logToEdit = allLogs[logTitle]
				logToEdit.SetText(logNewText)
				logToEdit.SetEditorTrue(*editor)
				logToEdit.SetNewEditedTimeNow()

				/* Reply to user */
				sendReply(
					bot,
					fmt.Sprintf("Success! <b>%s</b> has been edited by <b>[@%s]</b>:\n\n%s", logTitle, editor.UserName, logNewText),
					update.Message.Chat.ID,
					update.Message.MessageID)

			case "read":
				tokens := strings.Split(strings.TrimSpace(update.Message.CommandArguments()), " ")

				/* Check usage */
				if len(tokens) != 1 {
					reply := "Error: Wrong usage.\n\n<b>Sample Usage:</b>\n/read LogName"
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				logTitle := tokens[0]

				/* Check if the Log exists */
				if _, ok := allLogs[logTitle]; !ok {
					reply := fmt.Sprintf("Error: No Log with name <b>%s</b> exists!", logTitle)
					sendReply(bot, reply, update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				/* Reply to user */
				sendReply(
					bot,
					fmt.Sprintf("Success! Reading <b>%s</b>:\n\n%s\n\n<i>Last Edited at %s</i>", logTitle, allLogs[logTitle].Text, formatDate(allLogs[logTitle].LastEdited)),
					update.Message.Chat.ID,
					update.Message.MessageID)
			default:

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

func newLog(title string, creator tgbotapi.User) *Log {
	l := new(Log)
	l.Title = title
	l.Creator = creator

	l.UUID = uuid.New()
	l.Editors = make(map[tgbotapi.User]bool)
	l.Editors[creator] = true
	l.Text = ""
	l.CreatedAt = time.Now()
	l.LastEdited = l.CreatedAt
	return l
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
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}
