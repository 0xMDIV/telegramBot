package admin

import (
	"encoding/json"
	"log"
	"os"
	"telegramBot/config"
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BootstrapHandler macht den ersten User automatisch zum Admin
type BootstrapHandler struct{}

func NewBootstrapHandler() *BootstrapHandler {
	return &BootstrapHandler{}
}

func (h *BootstrapHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	// Nur bei DM-Nachrichten
	if update.Message == nil || update.Message.Chat.Type != "private" {
		return nil
	}

	// Nur wenn noch keine Admins vorhanden sind
	if len(b.GetConfig().Admin.AdminUserIDs) > 0 {
		return nil
	}

	// Ersten User als Admin hinzufügen
	userID := update.Message.From.ID
	username := bot.GetUserIdentifier(update.Message.From)

	if err := h.addFirstAdmin(userID); err != nil {
		log.Printf("Failed to bootstrap first admin: %v", err)
		return err
	}

	// Config neu laden
	newConfig, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Printf("Failed to reload config after bootstrap: %v", err)
		return err
	}

	*b.GetConfig() = *newConfig

	log.Printf("🚀 Bootstrap: First admin added - %d (%s)", userID, username)

	// Willkommensnachricht senden
	welcomeMsg := `🎉 Willkommen als erster Bot-Administrator!

Du wurdest automatisch als Admin hinzugefügt, da noch keine Admins konfiguriert waren.

💡 Was du jetzt tun kannst:
• /config - Bot-Einstellungen anpassen
• /help - Alle verfügbaren Commands anzeigen
• /add_admin @user - Weitere Admins hinzufügen

✅ Du hast jetzt volle Bot-Administrator-Rechte!`

	b.SendMessage(update.Message.Chat.ID, welcomeMsg)
	return nil
}

func (h *BootstrapHandler) addFirstAdmin(userID int64) error {
	configPath := "config/config.json"

	// Config laden
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Admin-Array holen
	admin, ok := cfg["admin"].(map[string]interface{})
	if !ok {
		return nil
	}

	// User als ersten Admin hinzufügen
	admin["admin_user_ids"] = []interface{}{userID}

	// Config speichern
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}
