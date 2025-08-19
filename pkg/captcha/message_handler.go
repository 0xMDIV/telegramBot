package captcha

import (
	"fmt"
	"strconv"
	"strings"
	"telegramBot/pkg/bot"
	"telegramBot/pkg/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

func (h *MessageHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil || update.Message.From == nil {
		return nil
	}

	// Nur Gruppen-Nachrichten verarbeiten
	if update.Message.Chat.Type == "private" {
		return nil
	}

	// Prüfen ob User pending Captcha hat
	pendingUser, err := b.GetDB().GetPendingUser(update.Message.From.ID, update.Message.Chat.ID)
	if err != nil || pendingUser == nil {
		return nil // User hat kein pending Captcha
	}

	// Prüfen ob Captcha abgelaufen ist
	if time.Now().After(pendingUser.ExpiresAt) {
		b.GetDB().RemovePendingUser(update.Message.From.ID, update.Message.Chat.ID)
		return nil
	}

	// User-Nachricht als potentielle Captcha-Antwort behandeln
	return h.handleCaptchaResponse(b, update, pendingUser)
}

func (h *MessageHandler) handleCaptchaResponse(b *bot.Bot, update tgbotapi.Update, pendingUser *database.PendingUser) error {
	userAnswer := strings.TrimSpace(update.Message.Text)

	// Prüfen ob es eine Zahl ist
	answerInt, err := strconv.Atoi(userAnswer)
	if err != nil {
		// Keine Zahl - User-Nachricht löschen
		b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
		return nil
	}

	// Korrekte Antwort berechnen
	correctAnswer, err := solveCaptcha(pendingUser.CaptchaKey)
	if err != nil {
		return fmt.Errorf("failed to solve captcha: %w", err)
	}

	if answerInt == correctAnswer {
		return h.handleCorrectCaptchaAnswer(b, update, pendingUser)
	} else {
		return h.handleWrongCaptchaAnswer(b, update, pendingUser)
	}
}

func (h *MessageHandler) handleCorrectCaptchaAnswer(b *bot.Bot, update tgbotapi.Update, pendingUser *database.PendingUser) error {
	// User freischalten - normale User-Rechte geben
	permissions := tgbotapi.ChatPermissions{
		CanSendMessages:       true,
		CanSendMediaMessages:  true,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
		CanAddWebPagePreviews: true,
		CanChangeInfo:         true, // Normale User können Chat-Info ändern
		CanInviteUsers:        true, // Normale User können andere einladen
		CanPinMessages:        true, // Normale User können Nachrichten pinnen
	}

	if err := b.RestrictChatMember(update.Message.Chat.ID, update.Message.From.ID, permissions); err != nil {
		return fmt.Errorf("failed to unrestrict user: %w", err)
	}

	// Willkommensnachricht löschen (falls vorhanden)
	welcomeMessageID, err := b.GetDB().GetWelcomeMessage(update.Message.From.ID, update.Message.Chat.ID)
	if err == nil && welcomeMessageID > 0 {
		b.DeleteMessage(update.Message.Chat.ID, welcomeMessageID)
		b.GetDB().RemoveWelcomeMessage(update.Message.From.ID, update.Message.Chat.ID)
	}

	// Log erfolgreiche Captcha-Lösung
	username := bot.GetUserIdentifier(update.Message.From)
	b.GetEventLogger().LogCaptchaSuccess(update.Message.Chat.ID, update.Message.From.ID, username, pendingUser.Attempts+1)

	// User aus Pending-Liste entfernen
	if err := b.GetDB().RemovePendingUser(update.Message.From.ID, update.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to remove pending user: %w", err)
	}

	// User-Antwort löschen
	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)

	// Erfolgs-Nachricht senden
	successMsg, err := b.SendMessage(update.Message.Chat.ID, fmt.Sprintf(
		"✅ %s hat das Captcha erfolgreich gelöst!",
		bot.GetUserMention(update.Message.From),
	))

	if err == nil {
		// Erfolgs-Nachricht nach separatem konfigurierten Delay löschen
		go func() {
			delay := time.Duration(b.GetConfig().Captcha.SuccessMessageDeleteDelayMinutes) * time.Minute
			time.Sleep(delay)
			b.DeleteMessage(update.Message.Chat.ID, successMsg.MessageID)
		}()
	}

	return nil
}

func (h *MessageHandler) handleWrongCaptchaAnswer(b *bot.Bot, update tgbotapi.Update, pendingUser *database.PendingUser) error {
	// Versuchsanzahl erhöhen
	if err := b.GetDB().IncrementAttempts(update.Message.From.ID, update.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	// User-Antwort löschen
	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)

	attempts := pendingUser.Attempts + 1
	maxAttempts := b.GetConfig().Captcha.MaxAttempts

	if attempts >= maxAttempts {
		// Maximale Versuche erreicht - User kicken
		username := bot.GetUserIdentifier(update.Message.From)
		b.GetEventLogger().LogCaptchaFail(update.Message.Chat.ID, update.Message.From.ID, username, "Too many wrong attempts")
		b.GetEventLogger().LogKick(update.Message.Chat.ID, update.Message.From.ID, username, "Captcha failed - too many attempts")

		if err := b.GetDB().RemovePendingUser(update.Message.From.ID, update.Message.Chat.ID); err != nil {
			return fmt.Errorf("failed to remove pending user: %w", err)
		}

		b.KickChatMember(update.Message.Chat.ID, update.Message.From.ID)
		b.UnbanChatMember(update.Message.Chat.ID, update.Message.From.ID)

		// Kick-Nachricht senden
		kickMsg, _ := b.SendMessage(update.Message.Chat.ID, fmt.Sprintf(
			"❌ %s wurde wegen zu vieler falscher Captcha-Versuche aus der Gruppe entfernt.",
			bot.GetUserMention(update.Message.From),
		))

		// Kick-Nachricht nach 5 Sekunden löschen
		if kickMsg.MessageID != 0 {
			go func() {
				time.Sleep(5 * time.Second)
				b.DeleteMessage(update.Message.Chat.ID, kickMsg.MessageID)
			}()
		}
	} else {
		// Noch Versuche übrig - Warnung senden
		remainingAttempts := maxAttempts - attempts
		warningMsg, _ := b.SendMessage(update.Message.Chat.ID, fmt.Sprintf(
			"❌ %s: Falsche Antwort! Noch %d Versuche übrig.",
			bot.GetUserMention(update.Message.From),
			remainingAttempts,
		))

		// Warnung nach 3 Sekunden löschen
		if warningMsg.MessageID != 0 {
			go func() {
				time.Sleep(3 * time.Second)
				b.DeleteMessage(update.Message.Chat.ID, warningMsg.MessageID)
			}()
		}
	}

	return nil
}
