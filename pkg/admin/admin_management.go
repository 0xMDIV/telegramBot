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

type AddAdminHandler struct{}

func NewAddAdminHandler() *AddAdminHandler {
	return &AddAdminHandler{}
}

func (h *AddAdminHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil || !update.Message.IsCommand() {
		return nil
	}

	// Prüfen ob User Admin ist (entweder in Config oder Gruppen-Admin)
	if !h.isUserAuthorized(b, update.Message.Chat.ID, update.Message.From.ID) {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.", 5)
		return err
	}

	args := update.Message.CommandArguments()
	if args == "" {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Verwendung: /add_admin <user_id> oder antworte auf eine Nachricht", 5)
		return err
	}

	var userID int64
	var err error

	// Prüfen ob auf eine Nachricht geantwortet wurde
	if update.Message.ReplyToMessage != nil {
		userID = update.Message.ReplyToMessage.From.ID
	} else {
		// User ID aus Argumenten parsen
		userID, err = strconv.ParseInt(strings.TrimSpace(args), 10, 64)
		if err != nil {
			_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Ungültige User ID", 5)
			return err
		}
	}

	// Admin hinzufügen
	if err := h.addAdminToConfig(userID); err != nil {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Fehler beim Hinzufügen des Admins", 5)
		return err
	}

	// Bot-Config neu laden
	newConfig, err := config.LoadConfig("config/config.json")
	if err != nil {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Fehler beim Neuladen der Konfiguration", 5)
		return err
	}

	// Config im Bot aktualisieren
	*b.GetConfig() = *newConfig

	_, err = b.SendTemporaryGroupMessage(update.Message.Chat.ID, fmt.Sprintf("✅ User %d wurde als Admin hinzugefügt", userID), 5)
	return err
}

func (h *AddAdminHandler) isUserAuthorized(b *bot.Bot, chatID, userID int64) bool {
	// Prüfen ob User bereits Bot-Admin ist
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}

	// Prüfen ob User Gruppen-Admin ist (nur in Gruppen, nicht in DMs)
	if chatID < 0 { // Negative Chat-IDs sind Gruppen
		isAdmin, err := b.IsUserAdmin(chatID, userID)
		if err == nil && isAdmin {
			return true
		}
	}

	return false
}

func (h *AddAdminHandler) addAdminToConfig(userID int64) error {
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
			return fmt.Errorf("user is already admin")
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

type DelAdminHandler struct{}

func NewDelAdminHandler() *DelAdminHandler {
	return &DelAdminHandler{}
}

func (h *DelAdminHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil || !update.Message.IsCommand() {
		return nil
	}

	// Prüfen ob User Admin ist (entweder in Config oder Gruppen-Admin)
	if !h.isUserAuthorized(b, update.Message.Chat.ID, update.Message.From.ID) {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.", 5)
		return err
	}

	args := update.Message.CommandArguments()
	if args == "" {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Verwendung: /del_admin <user_id> oder antworte auf eine Nachricht", 5)
		return err
	}

	var userID int64
	var err error

	// Prüfen ob auf eine Nachricht geantwortet wurde
	if update.Message.ReplyToMessage != nil {
		userID = update.Message.ReplyToMessage.From.ID
	} else {
		// User ID aus Argumenten parsen
		userID, err = strconv.ParseInt(strings.TrimSpace(args), 10, 64)
		if err != nil {
			_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Ungültige User ID", 5)
			return err
		}
	}

	// Admin entfernen
	if err := h.removeAdminFromConfig(userID); err != nil {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Fehler beim Entfernen des Admins oder User ist kein Admin", 5)
		return err
	}

	// Bot-Config neu laden
	newConfig, err := config.LoadConfig("config/config.json")
	if err != nil {
		_, err := b.SendTemporaryGroupMessage(update.Message.Chat.ID, "❌ Fehler beim Neuladen der Konfiguration", 5)
		return err
	}

	// Config im Bot aktualisieren
	*b.GetConfig() = *newConfig

	_, err = b.SendTemporaryGroupMessage(update.Message.Chat.ID, fmt.Sprintf("✅ User %d wurde als Admin entfernt", userID), 5)
	return err
}

func (h *DelAdminHandler) isUserAuthorized(b *bot.Bot, chatID, userID int64) bool {
	// Prüfen ob User bereits Bot-Admin ist
	cfg := b.GetConfig()
	for _, adminID := range cfg.Admin.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}

	// Prüfen ob User Gruppen-Admin ist (nur in Gruppen, nicht in DMs)
	if chatID < 0 { // Negative Chat-IDs sind Gruppen
		isAdmin, err := b.IsUserAdmin(chatID, userID)
		if err == nil && isAdmin {
			return true
		}
	}

	return false
}

func (h *DelAdminHandler) removeAdminFromConfig(userID int64) error {
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
		return fmt.Errorf("no admins found")
	}

	// User entfernen
	newAdminIDs := []interface{}{}
	found := false
	for _, id := range adminIDs {
		if int64(id.(float64)) != userID {
			newAdminIDs = append(newAdminIDs, id)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user is not admin")
	}

	admin["admin_user_ids"] = newAdminIDs

	// Config speichern
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}
