package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"telegramBot/config"
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

func (h *ConfigHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil || !update.Message.IsCommand() {
		return nil
	}

	// Nur DM-Nachrichten verarbeiten
	if update.Message.Chat.Type != "private" {
		return nil
	}

	// Prüfen ob User Admin ist
	if !h.isUserAdmin(b, update.Message.From.ID) {
		_, err := b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return err
	}

	args := update.Message.CommandArguments()
	if args == "" {
		return h.showConfigMenu(b, update.Message.Chat.ID)
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return h.showConfigMenu(b, update.Message.Chat.ID)
	}

	key := parts[0]
	value := parts[1]

	return h.updateConfig(b, update.Message.Chat.ID, key, value)
}

func (h *ConfigHandler) isUserAdmin(b *bot.Bot, userID int64) bool {
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

func (h *ConfigHandler) showConfigMenu(b *bot.Bot, chatID int64) error {
	cfg := b.GetConfig()

	text := fmt.Sprintf(`⚙️ **Bot Konfiguration**

📋 **Verfügbare Konfigurationsschlüssel:**

🔒 **Captcha Einstellungen:**
• **timeout_minutes** = %d
  └─ Zeitlimit für Captcha in Minuten (1-60)

• **max_attempts** = %d
  └─ Maximale Versuche für Captcha (1-10)

• **welcome_message** = "%s"
  └─ Willkommensnachricht für neue User

• **message_delete_delay_minutes** = %d
  └─ Löschzeit für Willkommensnachrichten (1-60)

• **success_message_delete_delay_minutes** = %d
  └─ Löschzeit für Erfolgsnachrichten (1-60)

👑 **Admin Einstellungen:**
• **default_mute_hours** = %d
  └─ Standard Mute Dauer in Stunden (1-168)

• **max_delete_messages** = %d
  └─ Max löschbare Nachrichten pro Command (1-1000)

📝 **Verwendung:**
/config <schlüssel> <wert>

📌 **Beispiele:**
• /config timeout_minutes 10
• /config welcome_message "Hallo! Willkommen!"
• /config success_message_delete_delay_minutes 2
• /config max_attempts 5`,
		cfg.Captcha.TimeoutMinutes,
		cfg.Captcha.MaxAttempts,
		cfg.Captcha.WelcomeMessage,
		cfg.Captcha.MessageDeleteDelayMinutes,
		cfg.Captcha.SuccessMessageDeleteDelayMinutes,
		cfg.Admin.DefaultMuteHours,
		cfg.Admin.MaxDeleteMessages)

	_, err := b.SendMessage(chatID, text)
	return err
}

func (h *ConfigHandler) updateConfig(b *bot.Bot, chatID int64, key, value string) error {
	configPath := "config/config.json"

	// Config laden
	data, err := os.ReadFile(configPath)
	if err != nil {
		b.SendMessage(chatID, "❌ Fehler beim Lesen der Konfigurationsdatei.")
		return err
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		b.SendMessage(chatID, "❌ Fehler beim Parsen der Konfigurationsdatei.")
		return err
	}

	// Config-Wert aktualisieren
	success := false
	switch key {
	case "timeout_minutes":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if captcha, ok := cfg["captcha"].(map[string]interface{}); ok {
				captcha["timeout_minutes"] = val
				success = true
			}
		}
	case "max_attempts":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if captcha, ok := cfg["captcha"].(map[string]interface{}); ok {
				captcha["max_attempts"] = val
				success = true
			}
		}
	case "welcome_message":
		if captcha, ok := cfg["captcha"].(map[string]interface{}); ok {
			captcha["welcome_message"] = value
			success = true
		}
	case "message_delete_delay_minutes":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if captcha, ok := cfg["captcha"].(map[string]interface{}); ok {
				captcha["message_delete_delay_minutes"] = val
				success = true
			}
		}
	case "success_message_delete_delay_minutes":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if captcha, ok := cfg["captcha"].(map[string]interface{}); ok {
				captcha["success_message_delete_delay_minutes"] = val
				success = true
			}
		}
	case "default_mute_hours":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if admin, ok := cfg["admin"].(map[string]interface{}); ok {
				admin["default_mute_hours"] = val
				success = true
			}
		}
	case "max_delete_messages":
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			if admin, ok := cfg["admin"].(map[string]interface{}); ok {
				admin["max_delete_messages"] = val
				success = true
			}
		}
	default:
		b.SendMessage(chatID, fmt.Sprintf("❌ Unbekannter Konfigurationsschlüssel: %s", key))
		return nil
	}

	if !success {
		b.SendMessage(chatID, "❌ Ungültiger Wert für den angegebenen Schlüssel.")
		return nil
	}

	// Config speichern
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		b.SendMessage(chatID, "❌ Fehler beim Erstellen der neuen Konfiguration.")
		return err
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		b.SendMessage(chatID, "❌ Fehler beim Speichern der Konfigurationsdatei.")
		return err
	}

	// Bot-Config neu laden
	newConfig, err := config.LoadConfig(configPath)
	if err != nil {
		b.SendMessage(chatID, "❌ Fehler beim Neuladen der Konfiguration.")
		return err
	}

	// Config im Bot aktualisieren
	*b.GetConfig() = *newConfig

	b.SendMessage(chatID, fmt.Sprintf("✅ Konfiguration erfolgreich aktualisiert!\n%s = %s", key, value))
	return nil
}
