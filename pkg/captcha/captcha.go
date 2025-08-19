package captcha

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"telegramBot/pkg/bot"
	"telegramBot/pkg/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil || update.Message.NewChatMembers == nil {
		return nil
	}

	chatID := update.Message.Chat.ID
	if update.Message.Chat.Type == "private" {
		return nil
	}

	for _, newMember := range update.Message.NewChatMembers {
		if newMember.IsBot {
			continue
		}

		if err := h.handleNewMember(b, chatID, &newMember); err != nil {
			return fmt.Errorf("failed to handle new member %d: %w", newMember.ID, err)
		}
	}

	return nil
}

func (h *Handler) handleNewMember(b *bot.Bot, chatID int64, user *tgbotapi.User) error {
	permissions := tgbotapi.ChatPermissions{
		CanSendMessages:       true, // User darf Nachrichten senden für Captcha-Antworten
		CanSendMediaMessages:  false,
		CanSendPolls:          false,
		CanSendOtherMessages:  false,
		CanAddWebPagePreviews: false,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
	}

	if err := b.RestrictChatMember(chatID, user.ID, permissions); err != nil {
		return fmt.Errorf("failed to restrict user: %w", err)
	}

	captchaKey := generateCaptcha()
	solution, _ := solveCaptcha(captchaKey)

	pendingUser := database.PendingUser{
		UserID:     user.ID,
		ChatID:     chatID,
		CaptchaKey: captchaKey,
		ExpiresAt:  time.Now().Add(time.Duration(b.GetConfig().Captcha.MessageDeleteDelayMinutes) * time.Minute),
		Attempts:   0,
	}

	if err := b.GetDB().AddPendingUser(pendingUser); err != nil {
		return fmt.Errorf("failed to add pending user: %w", err)
	}

	return h.sendCaptchaToGroup(b, user, captchaKey, solution, chatID)
}

func (h *Handler) sendCaptchaToGroup(b *bot.Bot, user *tgbotapi.User, captchaKey string, solution int, chatID int64) error {
	text := fmt.Sprintf(
		"%s %s!\n\n"+
			"Um in der Gruppe schreiben zu können, löse bitte das folgende Captcha:\n\n"+
			"Berechne: %s = ?\n\n"+
			"Antworte einfach mit der Zahl. Du hast %d Minuten Zeit.",
		b.GetConfig().Captcha.WelcomeMessage,
		bot.GetUserMention(user),
		captchaKey,
		b.GetConfig().Captcha.MessageDeleteDelayMinutes,
	)

	// Willkommensnachricht mit Captcha senden
	welcomeMsg, err := b.SendMessage(chatID, text)
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	// Willkommensnachricht-ID in der DB speichern für spätere Löschung
	if err := b.GetDB().SetWelcomeMessage(user.ID, chatID, welcomeMsg.MessageID); err != nil {
		// Fallback: Direkt Timer für Willkommensnachricht starten
		go func() {
			delay := time.Duration(b.GetConfig().Captcha.MessageDeleteDelayMinutes) * time.Minute
			time.Sleep(delay)
			b.DeleteMessage(chatID, welcomeMsg.MessageID)
		}()
	}

	// Auto-Kick Timer starten
	go func() {
		delay := time.Duration(b.GetConfig().Captcha.MessageDeleteDelayMinutes) * time.Minute
		time.Sleep(delay)

		// Prüfen ob User noch pending ist
		pendingUser, err := b.GetDB().GetPendingUser(user.ID, chatID)
		if err == nil && pendingUser != nil {
			// User hat Captcha nicht gelöst - kicken
			username := bot.GetUserIdentifier(user)
			b.GetEventLogger().LogCaptchaFail(chatID, user.ID, username, "Timeout - captcha not solved in time")
			b.GetEventLogger().LogKick(chatID, user.ID, username, "Captcha timeout")

			b.KickChatMember(chatID, user.ID)
			b.UnbanChatMember(chatID, user.ID)
			b.GetDB().RemovePendingUser(user.ID, chatID)

			// Willkommensnachricht löschen
			b.DeleteMessage(chatID, welcomeMsg.MessageID)

			// Timeout-Nachricht senden
			timeoutMsg, _ := b.SendMessage(chatID, fmt.Sprintf(
				"%s wurde wegen Captcha-Timeout aus der Gruppe entfernt.",
				bot.GetUserMention(user),
			))

			// Timeout-Nachricht nach 5 Sekunden löschen
			go func() {
				time.Sleep(5 * time.Second)
				b.DeleteMessage(chatID, timeoutMsg.MessageID)
			}()
		}
	}()

	return nil

}

func generateCaptcha() string {
	rand.Seed(time.Now().UnixNano())
	a := rand.Intn(20) + 1
	b := rand.Intn(20) + 1
	return fmt.Sprintf("%d+%d", a, b)
}

func generateMathProblem(captcha string) string {
	return captcha + " = ?"
}

func solveCaptcha(captcha string) (int, error) {
	parts := strings.Split(captcha, "+")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid captcha format")
	}

	a, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, err
	}

	b, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, err
	}

	return a + b, nil
}

type CallbackHandler struct{}

func NewCallbackHandler() *CallbackHandler {
	return &CallbackHandler{}
}

func (h *CallbackHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.CallbackQuery == nil {
		return nil
	}

	callback := update.CallbackQuery
	data := callback.Data

	if strings.HasPrefix(data, "captcha_solve:") {
		return h.handleCaptchaSolve(b, callback)
	}

	if strings.HasPrefix(data, "captcha_answer:") {
		return h.handleCaptchaAnswer(b, callback)
	}

	return nil
}

func (h *CallbackHandler) handleCaptchaSolve(b *bot.Bot, callback *tgbotapi.CallbackQuery) error {
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid callback data format")
	}

	groupChatID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid group chat ID")
	}

	captchaKey := parts[2]

	pendingUser, err := b.GetDB().GetPendingUser(callback.From.ID, groupChatID)
	if err != nil {
		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Captcha nicht gefunden oder abgelaufen!"))
		return nil
	}

	if pendingUser.CaptchaKey != captchaKey {
		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Ungültiges Captcha!"))
		return nil
	}

	if time.Now().After(pendingUser.ExpiresAt) {
		b.GetDB().RemovePendingUser(callback.From.ID, groupChatID)
		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Captcha abgelaufen!"))
		return nil
	}

	solution, err := solveCaptcha(captchaKey)
	if err != nil {
		return fmt.Errorf("failed to solve captcha: %w", err)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	correctAnswer := solution
	wrongAnswers := []int{correctAnswer + 1, correctAnswer - 1, correctAnswer + 2}

	answers := append([]int{correctAnswer}, wrongAnswers...)
	rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })

	for _, answer := range answers {
		button := tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(answer),
			fmt.Sprintf("captcha_answer:%d:%s:%d", groupChatID, captchaKey, answer),
		)
		row = append(row, button)
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	text := fmt.Sprintf(
		"Captcha-Loesung\n\n"+
			"Berechne: %s\n\n"+
			"Wähle die richtige Antwort:",
		generateMathProblem(captchaKey),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ReplyMarkup = &keyboard
	edit.ParseMode = "Markdown"

	_, err = b.GetAPI().Send(edit)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, ""))
	return nil
}

func (h *CallbackHandler) handleCaptchaAnswer(b *bot.Bot, callback *tgbotapi.CallbackQuery) error {
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 4 {
		return fmt.Errorf("invalid callback data format")
	}

	groupChatID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid group chat ID")
	}

	captchaKey := parts[2]
	userAnswer, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("invalid answer format")
	}

	pendingUser, err := b.GetDB().GetPendingUser(callback.From.ID, groupChatID)
	if err != nil {
		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Captcha nicht gefunden!"))
		return nil
	}

	if time.Now().After(pendingUser.ExpiresAt) {
		b.GetDB().RemovePendingUser(callback.From.ID, groupChatID)
		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Captcha abgelaufen!"))
		return nil
	}

	correctAnswer, err := solveCaptcha(captchaKey)
	if err != nil {
		return fmt.Errorf("failed to solve captcha: %w", err)
	}

	if userAnswer == correctAnswer {
		return h.handleCorrectAnswer(b, callback, groupChatID)
	} else {
		return h.handleWrongAnswer(b, callback, pendingUser, groupChatID)
	}
}

func (h *CallbackHandler) handleCorrectAnswer(b *bot.Bot, callback *tgbotapi.CallbackQuery, groupChatID int64) error {
	permissions := tgbotapi.ChatPermissions{
		CanSendMessages:       true,
		CanSendMediaMessages:  true,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
		CanAddWebPagePreviews: true,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
	}

	if err := b.RestrictChatMember(groupChatID, callback.From.ID, permissions); err != nil {
		return fmt.Errorf("failed to unrestrict user: %w", err)
	}

	if err := b.GetDB().RemovePendingUser(callback.From.ID, groupChatID); err != nil {
		return fmt.Errorf("failed to remove pending user: %w", err)
	}

	successText := "Glückwunsch!\n\nDu hast das Captcha erfolgreich gelöst und wurdest zur Gruppe hinzugefügt!"
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, successText)
	edit.ParseMode = "Markdown"
	b.GetAPI().Send(edit)

	welcomeText := fmt.Sprintf(
		"%s %s",
		b.GetConfig().Captcha.WelcomeMessage,
		bot.GetUserMention(callback.From),
	)

	msg, err := b.SendMessage(groupChatID, welcomeText)
	if err == nil {
		// Nach messageDeleteDelayMinutes löschen
		go func() {
			delay := time.Duration(b.GetConfig().Captcha.MessageDeleteDelayMinutes) * time.Minute
			time.Sleep(delay)
			b.DeleteMessage(groupChatID, msg.MessageID)
		}()
	}

	// Auch die Captcha-Nachricht nach dem konfigurierten Zeitraum löschen
	go func() {
		delay := time.Duration(b.GetConfig().Captcha.MessageDeleteDelayMinutes) * time.Minute
		time.Sleep(delay)
		b.DeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	}()

	b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "✅ Captcha gelöst!"))
	return nil
}

func (h *CallbackHandler) handleWrongAnswer(b *bot.Bot, callback *tgbotapi.CallbackQuery, pendingUser *database.PendingUser, groupChatID int64) error {
	if err := b.GetDB().IncrementAttempts(callback.From.ID, groupChatID); err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	attempts := pendingUser.Attempts + 1
	maxAttempts := b.GetConfig().Captcha.MaxAttempts

	if attempts >= maxAttempts {
		if err := b.GetDB().RemovePendingUser(callback.From.ID, groupChatID); err != nil {
			return fmt.Errorf("failed to remove pending user: %w", err)
		}

		failText := "Captcha fehlgeschlagen!\n\nDu hast zu viele falsche Versuche gemacht und wurdest aus der Gruppe entfernt."
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, failText)
		edit.ParseMode = "Markdown"
		b.GetAPI().Send(edit)

		// User aus Gruppe kicken
		b.KickChatMember(groupChatID, callback.From.ID)
		b.UnbanChatMember(groupChatID, callback.From.ID)

		// Captcha-Nachricht sofort nach Kick löschen
		go func() {
			time.Sleep(2 * time.Second) // Kurz warten, damit User die Nachricht sehen kann
			b.DeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		}()

		b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, "Zu viele Fehlversuche!"))
		return nil
	}

	remainingAttempts := maxAttempts - attempts
	retryText := fmt.Sprintf(
		"Falsche Antwort!\n\n"+
			"Du hast noch %d Versuche uebrig.\n\n"+
			"Klicke erneut auf 'Antwort eingeben' um es nochmal zu versuchen.",
		remainingAttempts,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Erneut versuchen", fmt.Sprintf("captcha_solve:%d:%s", groupChatID, pendingUser.CaptchaKey)),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, retryText)
	edit.ReplyMarkup = &keyboard
	edit.ParseMode = "Markdown"
	b.GetAPI().Send(edit)

	b.GetAPI().Send(tgbotapi.NewCallback(callback.ID, fmt.Sprintf("Falsch! Noch %d Versuche", remainingAttempts)))
	return nil
}
