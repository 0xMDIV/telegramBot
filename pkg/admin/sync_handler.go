package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"telegramBot/config"
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SyncAdminsHandler wird bei jeder Admin-Aktion aufgerufen um Gruppen-Admins zu synchronisieren
type SyncAdminsHandler struct{}

func NewSyncAdminsHandler() *SyncAdminsHandler {
	return &SyncAdminsHandler{}
}

// SyncGroupAdminsToBot synchronisiert alle Gruppen-Admins als Bot-Admins
func (h *SyncAdminsHandler) SyncGroupAdminsToBot(b *bot.Bot, chatID int64) error {
	// Alle Admins der Gruppe holen
	configAPI := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	}

	admins, err := b.GetAPI().GetChatAdministrators(configAPI)
	if err != nil {
		return fmt.Errorf("failed to get chat administrators: %w", err)
	}

	addedCount := 0

	// Für jeden Admin prüfen und ggf. als Bot-Admin hinzufügen
	for _, admin := range admins {
		if admin.User.IsBot {
			continue // Bots überspringen
		}

		if !h.isBotAdmin(b, admin.User.ID) {
			if err := h.addUserToBotAdmins(admin.User.ID); err != nil {
				log.Printf("Failed to add group admin %d as bot admin: %v", admin.User.ID, err)
			} else {
				username := bot.GetUserIdentifier(admin.User)
				log.Printf("Auto-Sync: Group admin %d (%s) wurde als Bot-Admin hinzugefügt",
					admin.User.ID, username)
				addedCount++
			}
		}
	}

	if addedCount > 0 {
		// Bot-Config neu laden
		newConfig, err := config.LoadConfig("config/config.json")
		if err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}

		// Config im Bot aktualisieren
		*b.GetConfig() = *newConfig

		log.Printf("Auto-Sync completed: %d new bot admins added from group %d", addedCount, chatID)
	}

	return nil
}

// CheckDMUserForGroupAdmin prüft bei DM-Nachrichten ob User in bekannten Gruppen Admin ist
func (h *SyncAdminsHandler) CheckDMUserForGroupAdmin(b *bot.Bot, userID int64) {
	// Wenn bereits Bot-Admin, nichts tun
	if h.isBotAdmin(b, userID) {
		return
	}

	// Da wir keine persistente Liste aller Gruppen haben, können wir hier nur loggen
	// Die tatsächliche Synchronisation erfolgt beim nächsten Admin-Command in einer Gruppe
	log.Printf("DM from user %d - will sync admin rights on next group interaction", userID)
}

func (h *SyncAdminsHandler) isBotAdmin(b *bot.Bot, userID int64) bool {
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

func (h *SyncAdminsHandler) addUserToBotAdmins(userID int64) error {
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
		return fmt.Errorf("admin section not found")
	}

	adminIDs, ok := admin["admin_user_ids"].([]interface{})
	if !ok {
		adminIDs = []interface{}{}
	}

	// Prüfen ob User bereits Admin ist
	for _, id := range adminIDs {
		if int64(id.(float64)) == userID {
			return nil // Bereits Admin
		}
	}

	// User hinzufügen
	adminIDs = append(adminIDs, userID)
	admin["admin_user_ids"] = adminIDs

	// Config speichern
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}
