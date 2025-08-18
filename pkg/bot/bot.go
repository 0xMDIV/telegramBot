package bot

import (
	"fmt"
	"log"
	"strings"
	"telegramBot/config"
	"telegramBot/pkg/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	config   *config.Config
	db       *database.DB
	handlers map[string]Handler
	logger   *CommandLogger
}

type Handler interface {
	Handle(bot *Bot, update tgbotapi.Update) error
}

func NewBot(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = cfg.Debug

	db, err := database.NewDB(cfg.Database.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger, err := NewCommandLogger("commands.log")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize command logger: %w", err)
	}

	bot := &Bot{
		api:      api,
		config:   cfg,
		db:       db,
		handlers: make(map[string]Handler),
		logger:   logger,
	}

	return bot, nil
}

func (b *Bot) RegisterHandler(command string, handler Handler) {
	b.handlers[command] = handler
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Printf("Bot %s started successfully", b.api.Self.UserName)

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in handler: %v", r)
		}
	}()

	if update.Message != nil {
		if update.Message.IsCommand() {
			command := update.Message.Command()
			if handler, exists := b.handlers[command]; exists {
				username := GetUserIdentifier(update.Message.From)
				args := update.Message.CommandArguments()

				// Für Gruppen-Commands: Lösche Command-Nachricht nach 5 Sekunden
				if update.Message.Chat.Type != "private" {
					go func() {
						time.Sleep(5 * time.Second)
						b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
						log.Printf("DEBUG: Auto-deleted command message %d", update.Message.MessageID)
					}()
				}

				err := handler.Handle(b, update)
				result := "SUCCESS"
				if err != nil {
					result = "ERROR: " + err.Error()
					log.Printf("Error handling command %s: %v", command, err)
				}

				b.logger.LogCommand(update.Message.Chat.ID, update.Message.From.ID, username, command, args, result)
				return
			}
		}

		if update.Message.NewChatMembers != nil {
			if handler, exists := b.handlers["new_member"]; exists {
				if err := handler.Handle(b, update); err != nil {
					log.Printf("Error handling new member: %v", err)
				}
			}
			return
		}

		if update.Message.LeftChatMember != nil {
			if handler, exists := b.handlers["left_member"]; exists {
				if err := handler.Handle(b, update); err != nil {
					log.Printf("Error handling left member: %v", err)
				}
			}
			return
		}

		if handler, exists := b.handlers["message"]; exists {
			if err := handler.Handle(b, update); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}

	if update.CallbackQuery != nil {
		if handler, exists := b.handlers["callback"]; exists {
			if err := handler.Handle(b, update); err != nil {
				log.Printf("Error handling callback: %v", err)
			}
		}
	}
}

func (b *Bot) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return b.api.Send(msg)
}

func (b *Bot) SendTemporaryGroupMessage(chatID int64, text string, deleteAfterSeconds int) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := b.api.Send(msg)

	if err == nil && deleteAfterSeconds > 0 {
		go func() {
			time.Sleep(time.Duration(deleteAfterSeconds) * time.Second)
			b.DeleteMessage(chatID, sentMsg.MessageID)
			log.Printf("DEBUG: Auto-deleted bot response message %d", sentMsg.MessageID)
		}()
	}

	return sentMsg, err
}

func (b *Bot) SendTemporaryMessage(chatID int64, text string, deleteAfterSeconds int) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := b.api.Send(msg)

	if err == nil && deleteAfterSeconds > 0 {
		go func() {
			time.Sleep(time.Duration(deleteAfterSeconds) * time.Second)
			b.DeleteMessage(chatID, sentMsg.MessageID)
		}()
	}

	return sentMsg, err
}

func (b *Bot) SendTemporaryMessageAndDeleteCommand(chatID int64, text string, commandMessageID int, deleteAfterSeconds int) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := b.api.Send(msg)

	if err == nil && deleteAfterSeconds > 0 {
		log.Printf("DEBUG: Scheduled deletion in %d seconds - Bot message ID: %d, Command message ID: %d",
			deleteAfterSeconds, sentMsg.MessageID, commandMessageID)

		go func() {
			time.Sleep(time.Duration(deleteAfterSeconds) * time.Second)

			log.Printf("DEBUG: Attempting to delete messages - Bot: %d, Command: %d", sentMsg.MessageID, commandMessageID)

			// Lösche sowohl Bot-Antwort als auch Command-Nachricht
			err1 := b.DeleteMessage(chatID, sentMsg.MessageID)
			err2 := b.DeleteMessage(chatID, commandMessageID)

			if err1 != nil {
				log.Printf("DEBUG: Failed to delete bot message %d: %v", sentMsg.MessageID, err1)
			} else {
				log.Printf("DEBUG: Successfully deleted bot message %d", sentMsg.MessageID)
			}

			if err2 != nil {
				log.Printf("DEBUG: Failed to delete command message %d: %v", commandMessageID, err2)
			} else {
				log.Printf("DEBUG: Successfully deleted command message %d", commandMessageID)
			}
		}()
	} else {
		log.Printf("DEBUG: No deletion scheduled - err: %v, deleteAfterSeconds: %d", err, deleteAfterSeconds)
	}

	return sentMsg, err
}

func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	return b.api.Send(msg)
}

func (b *Bot) DeleteMessage(chatID int64, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.api.Request(deleteMsg)
	return err
}

func (b *Bot) RestrictChatMember(chatID, userID int64, permissions tgbotapi.ChatPermissions) error {
	restrictConfig := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		Permissions: &permissions,
	}
	_, err := b.api.Request(restrictConfig)
	return err
}

func (b *Bot) KickChatMember(chatID, userID int64) error {
	banConfig := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
	}
	_, err := b.api.Request(banConfig)
	return err
}

func (b *Bot) BanChatMember(chatID, userID int64) error {
	banConfig := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
	}
	_, err := b.api.Request(banConfig)
	return err
}

func (b *Bot) UnbanChatMember(chatID, userID int64) error {
	unbanConfig := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		OnlyIfBanned: true,
	}
	_, err := b.api.Request(unbanConfig)
	return err
}

func (b *Bot) IsUserAdmin(chatID, userID int64) (bool, error) {
	chatMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false, err
	}

	status := chatMember.Status
	return status == "administrator" || status == "creator", nil
}

func (b *Bot) GetConfig() *config.Config {
	return b.config
}

func (b *Bot) GetDB() *database.DB {
	return b.db
}

func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) Stop() error {
	b.api.StopReceivingUpdates()
	if b.logger != nil {
		b.logger.Close()
	}
	return b.db.Close()
}

func GetUserMention(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}
	return fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.ID)
}

func FormatUserName(user *tgbotapi.User) string {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if user.UserName != "" {
		name += " (@" + user.UserName + ")"
	}
	return name
}

func ParseDuration(input string) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 1, nil
	}

	var hours int
	if _, err := fmt.Sscanf(input, "%d", &hours); err != nil {
		return 0, fmt.Errorf("invalid duration format")
	}

	if hours < 1 || hours > 24*7 {
		return 0, fmt.Errorf("duration must be between 1 hour and 7 days")
	}

	return hours, nil
}
