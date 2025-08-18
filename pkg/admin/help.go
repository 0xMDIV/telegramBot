package admin

import (
	"fmt"
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpHandler struct{}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

func (h *HelpHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return h.sendHelpDM(b, update.Message.From.ID)
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return err
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	if err := h.sendHelpDM(b, update.Message.From.ID); err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Ich konnte dir keine private Nachricht senden. Starte zuerst eine Unterhaltung mit mir.", 5)
		return nil
	}

	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
	return nil
}

func (h *HelpHandler) sendHelpDM(b *bot.Bot, userID int64) error {
	helpText := fmt.Sprintf(`Telegram Security Bot - Admin Hilfe

Verfuegbare Commands:

Moderation:
• /ban @user - User permanent bannen
• /kick @user - User aus Gruppe entfernen  
• /mute @user [Stunden] - User temporär muten (Standard: 1h)
• /unmute @user - Mute aufheben

Nachrichten:
• /del [Anzahl] - Letzten X Nachrichten loeschen (max. 100)
• /permissions - Bot-Rechte ueberpruefen

Verwendung:
• Als Antwort auf Nachricht: /ban, /kick, /mute 2
• Mit User-ID: /ban 123456789

WICHTIG: Username-Aufloesung funktioniert nur bei 'Auf Nachricht antworten'.
Direkte @username Eingabe wird nicht unterstuetzt.

Captcha-System:
Neue Mitglieder werden automatisch stummgeschaltet und müssen ein Captcha per DM lösen.

Konfiguration:
• Captcha-Timeout: %d Minuten
• Max. Versuche: %d
• Standard Mute-Dauer: %d Stunden
• Max. löschbare Nachrichten: %d

Support:
Bei Fragen oder Problemen wende dich an den Bot-Administrator.`,
		b.GetConfig().Captcha.TimeoutMinutes,
		b.GetConfig().Captcha.MaxAttempts,
		b.GetConfig().Admin.DefaultMuteHours,
		b.GetConfig().Admin.MaxDeleteMessages)

	_, err := b.SendMessage(userID, helpText)
	return err
}
