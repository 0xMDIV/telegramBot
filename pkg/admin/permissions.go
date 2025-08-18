package admin

import (
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PermissionsHandler struct{}

func NewPermissionsHandler() *PermissionsHandler {
	return &PermissionsHandler{}
}

func (h *PermissionsHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return err
	}

	if !isAdmin {
		_, _ = b.SendMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	hasPerms, status, err := b.CheckRequiredPermissions(update.Message.Chat.ID)
	if err != nil {
		_, _ = b.SendMessage(update.Message.Chat.ID, "Fehler beim Überprüfen der Bot-Rechte: "+err.Error())
		return err
	}

	if err := h.sendPermissionsDM(b, update.Message.From.ID, status, hasPerms); err != nil {
		_, _ = b.SendMessage(update.Message.Chat.ID, "Ich konnte dir keine private Nachricht senden. Starte zuerst eine Unterhaltung mit mir.")
		return nil
	}

	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
	return nil
}

func (h *PermissionsHandler) sendPermissionsDM(b *bot.Bot, userID int64, status string, hasAllPerms bool) error {
	message := status + "\n\n"

	if !hasAllPerms {
		message += "WARNUNG: Dem Bot fehlen wichtige Berechtigungen!\n\n"
		message += "So aktivierst du die Berechtigungen:\n"
		message += "1. Gehe zu den Gruppeneinstellungen\n"
		message += "2. Waehle 'Administratoren'\n"
		message += "3. Waehle den Bot aus\n"
		message += "4. Aktiviere die fehlenden Rechte\n\n"
		message += "Ohne diese Rechte funktionieren Commands wie /ban, /kick und /mute nicht!"
	} else {
		message += "Der Bot ist korrekt konfiguriert und einsatzbereit."
	}

	_, err := b.SendMessage(userID, message)
	return err
}
