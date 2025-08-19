package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"telegramBot/pkg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AutoAdminHandler struct{}

func NewAutoAdminHandler() *AutoAdminHandler {
	return &AutoAdminHandler{}
}

func (h *AutoAdminHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	// Nur DM-Nachrichten verarbeiten
	if update.Message == nil || update.Message.Chat.Type != "private" {
		return nil
	}

	// Prüfen ob User bereits Bot-Admin ist
	if h.isBotAdmin(b, update.Message.From.ID) {
		return nil // Bereits Bot-Admin
	}

	// Alle Gruppen durchgehen und prüfen ob User dort Admin ist
	return h.checkAndGrantAdminRights(b, update.Message.From.ID)
}

func (h *AutoAdminHandler) isBotAdmin(b *bot.Bot, userID int64) bool {
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

func (h *AutoAdminHandler) checkAndGrantAdminRights(b *bot.Bot, userID int64) error {
	// Alle aktiven Chats durchgehen (hier müssten wir eine Chat-Liste haben)
	// Da wir keine Chat-Liste haben, prüfen wir alle bekannten Gruppen aus der DB

	// Alternative: Bei der ersten DM eines Users prüfen wir alle Gruppen
	// Dies ist ein vereinfachter Ansatz - in der Praxis würde man die Chat-IDs speichern

	isGroupAdmin := h.checkIfUserIsGroupAdmin(b, userID)

	if isGroupAdmin {
		return h.addUserToBotAdmins(userID)
	}

	return nil
}

func (h *AutoAdminHandler) checkIfUserIsGroupAdmin(b *bot.Bot, userID int64) bool {
	// Da wir keine Liste aller Gruppen haben, können wir das nur bei Commands prüfen
	// Diese Funktion wird bei DM-Kontakt aufgerufen
	// Für eine vollständige Implementierung müssten wir eine Gruppen-Liste in der DB speichern

	// Für jetzt return false - die Logik wird in anderen Handlern integriert
	return false
}

func (h *AutoAdminHandler) addUserToBotAdmins(userID int64) error {
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

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return err
	}

	log.Printf("Auto-Admin: User %d wurde automatisch als Bot-Admin hinzugefügt", userID)
	return nil
}

// SyncGroupAdmins synchronisiert Gruppen-Admins mit Bot-Admins
func (h *AutoAdminHandler) SyncGroupAdmins(b *bot.Bot, chatID int64) error {
	// Alle Admins der Gruppe holen
	config := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	}

	admins, err := b.GetAPI().GetChatAdministrators(config)
	if err != nil {
		return fmt.Errorf("failed to get chat administrators: %w", err)
	}

	// Für jeden Admin prüfen und ggf. als Bot-Admin hinzufügen
	for _, admin := range admins {
		if admin.User.IsBot {
			continue // Bots überspringen
		}

		if !h.isBotAdmin(b, admin.User.ID) {
			if err := h.addUserToBotAdmins(admin.User.ID); err != nil {
				log.Printf("Failed to add group admin %d as bot admin: %v", admin.User.ID, err)
			} else {
				log.Printf("Auto-Admin: Group admin %d (@%s) wurde als Bot-Admin hinzugefügt",
					admin.User.ID, admin.User.UserName)
			}
		}
	}

	// Bot-Config neu laden (wird vom aufrufenden Code gemacht)

	return nil
}
