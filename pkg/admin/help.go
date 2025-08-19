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
	// Prüfen ob User Bot-Admin ist für erweiterte Hilfe
	isBotAdmin := h.isBotAdmin(b, userID)

	helpText := fmt.Sprintf(`🛡️ **Telegram Security Bot - Hilfe**

📋 **Moderation Commands:**
• **/ban** @user [Grund] - User permanent bannen
• **/kick** @user [Grund] - User aus Gruppe entfernen  
• **/mute** @user [Stunden] [Grund] - User temporär muten (Standard: 1h)
• **/unmute** @user - Mute aufheben
• **/del** [Anzahl] - Letzten X Nachrichten löschen (max. %d)

👑 **Admin-Management:**
• **/add_admin** @user - User als Bot-Admin hinzufügen
• **/add_admin** 123456789 - User per ID als Bot-Admin hinzufügen
• **/del_admin** @user - Bot-Admin Rechte entfernen
• **/del_admin** 123456789 - Bot-Admin per ID entfernen

ℹ️ **Hilfsbefehle:**
• **/help** - Diese Hilfe anzeigen
• **/permissions** - Bot-Rechte überprüfen

📝 **Verwendung:**
• **Als Antwort auf Nachricht:** /ban, /kick, /mute 2 Störend
• **Mit User-ID:** /ban 123456789 Spam
• **Mit @Username:** /mute @user 2 (nur bei kleinen Gruppen)

🔒 **Captcha-System:**
Neue Mitglieder lösen Captcha **direkt in der Gruppe**:
• Mathematische Aufgaben (z.B. "5+3 = ?")
• %d Minuten Zeit, %d Versuche
• Bei Erfolg: Volle Berechtigung nach %d Min gelöscht
• Bei Fehlschlag: Automatischer Kick

👥 **Admin-System:**
• **Gruppen-Admins:** Automatisch alle Bot-Rechte in ihrer Gruppe
• **Bot-Admins:** Globale Rechte + Config-Zugriff per DM`,
		b.GetConfig().Admin.MaxDeleteMessages,
		b.GetConfig().Captcha.MessageDeleteDelayMinutes,
		b.GetConfig().Captcha.MaxAttempts,
		b.GetConfig().Captcha.SuccessMessageDeleteDelayMinutes)

	if isBotAdmin {
		helpText += fmt.Sprintf(`

⚙️ **Bot-Admin Commands (nur per DM):**
• **/config** - Alle Konfigurationsoptionen anzeigen
• **/config** <schlüssel> <wert> - Einstellung ändern

📊 **Verfügbare Config-Optionen:**
• **timeout_minutes** = %d (Captcha-Zeitlimit)
• **max_attempts** = %d (Captcha-Versuche)  
• **welcome_message** = "%s"
• **message_delete_delay_minutes** = %d
• **success_message_delete_delay_minutes** = %d
• **default_mute_hours** = %d
• **max_delete_messages** = %d

📌 **Config-Beispiele:**
• /config timeout_minutes 10
• /config welcome_message "Willkommen!"
• /config success_message_delete_delay_minutes 2`,
			b.GetConfig().Captcha.TimeoutMinutes,
			b.GetConfig().Captcha.MaxAttempts,
			b.GetConfig().Captcha.WelcomeMessage,
			b.GetConfig().Captcha.MessageDeleteDelayMinutes,
			b.GetConfig().Captcha.SuccessMessageDeleteDelayMinutes,
			b.GetConfig().Admin.DefaultMuteHours,
			b.GetConfig().Admin.MaxDeleteMessages)
	}

	helpText += `

💡 **Tipps:**
• Commands funktionieren in Gruppen und per Antwort auf Nachrichten
• Bot-Admins können Config per DM ändern
• Alle Aktionen werden geloggt (commands.log, events.log)

🚨 **Support:**
Bei Fragen oder Problemen wende dich an den Bot-Administrator.`

	_, err := b.SendMessage(userID, helpText)
	return err
}

func (h *HelpHandler) isBotAdmin(b *bot.Bot, userID int64) bool {
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}
