package handlers

import (
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

func (h *MessageHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	message := update.Message
	
	if message.Chat.Type == "private" {
		return nil
	}

	isMuted, err := b.GetDB().IsUserMuted(message.From.ID, message.Chat.ID)
	if err != nil {
		return err
	}

	if isMuted {
		return b.DeleteMessage(message.Chat.ID, message.MessageID)
	}

	return nil
}