package main

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/google/uuid"
)

type Log struct {
	Title      string
	UUID       uuid.UUID
	Creator    tgbotapi.User
	Editors    map[tgbotapi.User]bool
	Text       string
	CreatedAt  time.Time
	LastEdited time.Time
}

func (log *Log) SetText(text string) {
	log.Text = text
}

func (log *Log) SetEditorTrue(user tgbotapi.User) {
	log.Editors[user] = true
}

func (log *Log) SetNewEditedTimeNow() {
	log.LastEdited = time.Now()
}
