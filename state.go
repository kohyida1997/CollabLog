package main

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State struct {
	LogOwnerMap   map[int64]map[string]*Log
	EditedLogsMap map[int64]map[string]*Log
	AllLogs       map[string]*Log
}

func NewState() *State {

	toReturn := new(State)
	toReturn.LogOwnerMap = make(map[int64]map[string]*Log)
	toReturn.EditedLogsMap = make(map[int64]map[string]*Log)
	toReturn.AllLogs = make(map[string]*Log)
	return toReturn
}

func (s *State) MakeNewLog(update tgbotapi.Update) string {
	args := strings.TrimSpace(update.Message.CommandArguments())

	/* Disallow white spaces in Log name */
	if strings.Contains(args, " ") {
		return "Error: Log name <b>cannot contain white-spaces</b>"
	}

	/* Todo: HANDLE DUPLICATE LOG NAMES */

	/* Name of Log cannot be empty */
	if args == "" {
		errorMsg := "Error: name of Log cannot be empty.\n\n<b>Sample Usage:</b>\n/new MyLogName"
		return errorMsg
	}

	newLogToAdd := NewLog(args, *update.Message.From)

	/* Store as creator */
	if _, ok := s.LogOwnerMap[update.Message.From.ID]; !ok {
		s.LogOwnerMap[update.Message.From.ID] = map[string]*Log{}
	}
	s.LogOwnerMap[update.Message.From.ID][args] = newLogToAdd

	/* Store as editor */
	if _, ok := s.EditedLogsMap[update.Message.From.ID]; !ok {
		s.EditedLogsMap[update.Message.From.ID] = map[string]*Log{}
	}
	s.EditedLogsMap[update.Message.From.ID][args] = newLogToAdd

	s.AllLogs[args] = newLogToAdd
	reply := fmt.Sprintf("Success! Created new Log <b>[%s]</b>\n", args)

	return reply
}

func (s *State) GetCreatedLogs(update tgbotapi.Update) string {
	reply := fmt.Sprintf("<b>Logs created by @%s</b>:\n\n", update.Message.From.UserName)
	userID := update.Message.From.ID

	if _, ok := s.LogOwnerMap[userID]; !ok {
		reply += "<i>No Logs found</i>\n"
		return reply
	}

	if len(s.LogOwnerMap[userID]) == 0 {
		reply += "<i>No Logs found</i>\n"
		return reply
	}

	for key := range s.LogOwnerMap[userID] {
		reply += key + "\n"
	}

	return reply
}

func (s *State) EditLog(update tgbotapi.Update) string {
	tokens := strings.Split(strings.TrimSpace(update.Message.CommandArguments()), " ")

	/* Check usage */
	if len(tokens) < 2 {
		reply := "Error: Wrong usage.\n\n<b>Sample Usage:</b>\n/edit LogName NewText"
		return reply
	}

	logTitle := tokens[0]
	logNewText := strings.Join(tokens[1:], " ")
	editor := update.Message.From

	/* Check if the Log exists */
	if _, ok := s.AllLogs[logTitle]; !ok {
		reply := fmt.Sprintf("Error: No Log with name <b>%s</b> exists!", logTitle)
		return reply
	}

	/* Update Log */
	var logToEdit *Log
	logToEdit = s.AllLogs[logTitle]
	logToEdit.SetText(logNewText)
	logToEdit.SetEditorTrue(*editor)
	logToEdit.SetNewEditedTimeNow()

	return fmt.Sprintf("Success! <b>%s</b> has been edited by <b>[@%s]</b>:\n\n%s", logTitle, editor.UserName, logNewText)
}

func (s *State) ReadLog(update tgbotapi.Update) string {
	tokens := strings.Split(strings.TrimSpace(update.Message.CommandArguments()), " ")

	/* Check usage */
	if len(tokens) != 1 {
		reply := "Error: Wrong usage.\n\n<b>Sample Usage:</b>\n/read LogName"
		return reply
	}

	logTitle := tokens[0]

	/* Check if the Log exists */
	if _, ok := s.AllLogs[logTitle]; !ok {
		reply := fmt.Sprintf("Error: No Log with name <b>%s</b> exists!", logTitle)
		return reply
	}

	return fmt.Sprintf("Success! Reading <b>%s</b>:\n\n%s\n\n<i>Last Edited at %s</i>", logTitle, s.AllLogs[logTitle].Text, formatDate(s.AllLogs[logTitle].LastEdited))

}

func (s *State) DeleteLog(update tgbotapi.Update) string {
	args := strings.TrimSpace(update.Message.CommandArguments())

	/* Disallow white spaces in Log name */
	if strings.Contains(args, " ") {
		return "Error: Log name <b>cannot contain white-spaces</b>"
	}

	/* Todo: HANDLE DUPLICATE LOG NAMES */

	/* Name of Log cannot be empty */
	if args == "" {
		errorMsg := "Error: name of Log cannot be empty.\n\n<b>Sample Usage:</b>\n/new MyLogName"
		return errorMsg
	}

	/* Check if the Log exists */
	if _, ok := s.AllLogs[args]; !ok {
		reply := fmt.Sprintf("Error: No Log with name <b>%s</b> exists!", args)
		return reply
	}

	/* Only the creators (owners) of the Log can delete the Log! */
	senderID := update.Message.From.ID
	if _, ok := s.LogOwnerMap[senderID][args]; !ok {
		return fmt.Sprintf("Error: <b>@%s</b> not allowed to delete Log created by <b>@%s</b>",
			update.Message.From.UserName, s.AllLogs[args].Creator.UserName)
	}

	delete(s.LogOwnerMap[senderID], args)
	delete(s.AllLogs, args)
	delete(s.EditedLogsMap[senderID], args)

	return fmt.Sprintf("Success! Deleted Log <b>[%s]</b>", args)
}
