package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotPermissions struct {
	CanDeleteMessages  bool
	CanRestrictMembers bool
	CanPromoteMembers  bool
	CanChangeInfo      bool
	CanInviteUsers     bool
	CanPinMessages     bool
}

func (b *Bot) GetBotPermissions(chatID int64) (*BotPermissions, error) {
	botMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: b.api.Self.ID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bot member info: %w", err)
	}

	permissions := &BotPermissions{}

	if botMember.Status == "administrator" {
		permissions.CanDeleteMessages = botMember.CanDeleteMessages
		permissions.CanRestrictMembers = botMember.CanRestrictMembers
		permissions.CanPromoteMembers = botMember.CanPromoteMembers
		permissions.CanChangeInfo = botMember.CanChangeInfo
		permissions.CanInviteUsers = botMember.CanInviteUsers
		permissions.CanPinMessages = botMember.CanPinMessages
	} else if botMember.Status == "creator" {
		permissions.CanDeleteMessages = true
		permissions.CanRestrictMembers = true
		permissions.CanPromoteMembers = true
		permissions.CanChangeInfo = true
		permissions.CanInviteUsers = true
		permissions.CanPinMessages = true
	}

	return permissions, nil
}

func (b *Bot) CheckRequiredPermissions(chatID int64) (bool, string, error) {
	permissions, err := b.GetBotPermissions(chatID)
	if err != nil {
		return false, "", err
	}

	var missing []string
	var current []string

	// Required permissions for basic functionality
	if !permissions.CanDeleteMessages {
		missing = append(missing, "Nachrichten loeschen")
	} else {
		current = append(current, "Nachrichten loeschen")
	}

	if !permissions.CanRestrictMembers {
		missing = append(missing, "Mitglieder einschraenken")
	} else {
		current = append(current, "Mitglieder einschraenken")
	}

	// Optional but recommended permissions
	if permissions.CanInviteUsers {
		current = append(current, "Nutzer einladen")
	}

	if permissions.CanPinMessages {
		current = append(current, "Nachrichten anheften")
	}

	if permissions.CanChangeInfo {
		current = append(current, "Gruppeninfo aendern")
	}

	allPermissions := len(missing) == 0

	status := fmt.Sprintf("Bot-Berechtigung Status:\n\nVorhanden: %s\n", strings.Join(current, ", "))
	if len(missing) > 0 {
		status += fmt.Sprintf("Fehlend: %s\n", strings.Join(missing, ", "))
		status += "\nBitte gebe dem Bot folgende Admin-Rechte:\n"
		status += "- Nachrichten loeschen\n"
		status += "- Mitglieder bannen\n"
		status += "- Mitglieder einschraenken"
	} else {
		status += "\nAlle erforderlichen Berechtigungen sind vorhanden."
	}

	return allPermissions, status, nil
}
