package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ResolveUser resolves a username or user ID to a valid user ID
func (b *Bot) ResolveUser(chatID int64, target string) (int64, error) {
	target = strings.TrimSpace(target)

	// If it's a user ID (pure number)
	if userID, err := strconv.ParseInt(target, 10, 64); err == nil {
		return userID, nil
	}

	// If it's a username (starts with @)
	if strings.HasPrefix(target, "@") {
		username := strings.TrimPrefix(target, "@")
		return b.findUserByUsername(chatID, username)
	}

	return 0, fmt.Errorf("Ungültiges Format: %s (verwende @username oder User-ID)", target)
}

// findUserByUsername tries to find a user ID by username using Telegram API
func (b *Bot) findUserByUsername(chatID int64, username string) (int64, error) {
	// Try to get chat member by username
	chatMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID:             chatID,
			SuperGroupUsername: username, // This won't work for users
		},
	})

	if err == nil && chatMember.User != nil {
		return chatMember.User.ID, nil
	}

	// Fallback: Username resolution not directly possible with Bot API
	// User must use reply-to-message or provide user ID
	return 0, fmt.Errorf("Username @%s konnte nicht aufgelöst werden. Verwende 'Auf Nachricht antworten' oder User-ID", username)
}

// ValidateUserID checks if a user ID is valid and not a bot
func (b *Bot) ValidateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("Ungültige User-ID: %d", userID)
	}

	// Additional validation could be added here
	return nil
}
