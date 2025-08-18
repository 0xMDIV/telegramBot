package admin

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegramBot/pkg/bot"
	"telegramBot/pkg/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BanHandler struct{}
type KickHandler struct{}
type MuteHandler struct{}
type DeleteHandler struct{}

func NewBanHandler() *BanHandler {
	return &BanHandler{}
}

func NewKickHandler() *KickHandler {
	return &KickHandler{}
}

func NewMuteHandler() *MuteHandler {
	return &MuteHandler{}
}

func NewDeleteHandler() *DeleteHandler {
	return &DeleteHandler{}
}

func (h *BanHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	return h.handleBan(b, update, false)
}

func (h *BanHandler) handleBan(b *bot.Bot, update tgbotapi.Update, isTemporary bool) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	targetUser, reason, err := extractTargetUserAndReason(b, update.Message)
	if err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, ""+err.Error(), 5)
		return nil
	}

	if targetUser.ID == update.Message.From.ID {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du kannst dich nicht selbst bannen.", 5)
		return nil
	}

	if targetUser.ID != 0 {
		targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
		if err != nil {
			_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Überprüfen der User-Berechtigung.")
			return fmt.Errorf("failed to check target admin status: %w", err)
		}

		if targetIsAdmin {
			_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Admins können nicht gebannt werden.")
			return nil
		}
	}

	if err := b.BanChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Bannen des Users. Überprüfe die Bot-Rechte.")
		return fmt.Errorf("failed to ban user: %w", err)
	}

	successMsg := fmt.Sprintf(
		"User gebannt\n\n"+
			"User: %s\n"+
			"Admin: %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	if reason != "" {
		successMsg += fmt.Sprintf("\nGrund: %s", reason)
	}

	_, _ = b.SendTemporaryMessageAndDeleteCommand(update.Message.Chat.ID, successMsg, update.Message.MessageID, 5)
	return nil
}

func (h *KickHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	targetUser, reason, err := extractTargetUserAndReason(b, update.Message)
	if err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, ""+err.Error(), 5)
		return nil
	}

	if targetUser.ID == update.Message.From.ID {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du kannst dich nicht selbst kicken.")
		return nil
	}

	if targetUser.ID != 0 {
		targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
		if err != nil {
			_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Überprüfen der User-Berechtigung.")
			return fmt.Errorf("failed to check target admin status: %w", err)
		}

		if targetIsAdmin {
			_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Admins können nicht gekickt werden.")
			return nil
		}
	}

	if err := b.KickChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Kicken des Users. Überprüfe die Bot-Rechte.")
		return fmt.Errorf("failed to kick user: %w", err)
	}

	if err := b.UnbanChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		return fmt.Errorf("failed to unban user after kick: %w", err)
	}

	successMsg := fmt.Sprintf(
		"User gekickt\n\n"+
			"User: %s\n"+
			"Admin: %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	if reason != "" {
		successMsg += fmt.Sprintf("\nGrund: %s", reason)
	}

	_, _ = b.SendTemporaryMessageAndDeleteCommand(update.Message.Chat.ID, successMsg, update.Message.MessageID, 5)
	return nil
}

func (h *MuteHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	targetUser, duration, reason, err := h.parseTargetUserDurationAndReason(b, update.Message)
	if err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, ""+err.Error(), 5)
		return nil
	}

	if targetUser.ID == update.Message.From.ID {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du kannst dich nicht selbst muten.")
		return nil
	}

	targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
	if err != nil {
		return fmt.Errorf("failed to check target admin status: %w", err)
	}

	if targetIsAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Admins können nicht gemutet werden.")
		return nil
	}

	muteUntil := time.Now().Add(time.Duration(duration) * time.Hour)

	mutedUser := database.MutedUser{
		UserID: targetUser.ID,
		ChatID: update.Message.Chat.ID,
		Until:  muteUntil,
	}

	if err := b.GetDB().AddMutedUser(mutedUser); err != nil {
		return fmt.Errorf("failed to add muted user to database: %w", err)
	}

	permissions := tgbotapi.ChatPermissions{
		CanSendMessages:       false,
		CanSendMediaMessages:  false,
		CanSendPolls:          false,
		CanSendOtherMessages:  false,
		CanAddWebPagePreviews: false,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
	}

	if err := b.RestrictChatMember(update.Message.Chat.ID, targetUser.ID, permissions); err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Muten des Users.")
		return fmt.Errorf("failed to mute user: %w", err)
	}

	successMsg := fmt.Sprintf(
		"User gemutet\n\n"+
			"User: %s\n"+
			"Dauer: %d Stunden\n"+
			"Bis: %s\n"+
			"Admin: %s",
		bot.FormatUserName(targetUser),
		duration,
		muteUntil.Format("02.01.2006 15:04"),
		bot.GetUserMention(update.Message.From),
	)

	if reason != "" {
		successMsg += fmt.Sprintf("\nGrund: %s", reason)
	}

	_, _ = b.SendTemporaryMessageAndDeleteCommand(update.Message.Chat.ID, successMsg, update.Message.MessageID, 5)

	go func() {
		time.Sleep(time.Duration(duration) * time.Hour)
		h.unmuteUser(b, update.Message.Chat.ID, targetUser.ID)
	}()

	return nil
}

func (h *MuteHandler) parseTargetUserDurationAndReason(b *bot.Bot, message *tgbotapi.Message) (*tgbotapi.User, int, string, error) {
	args := strings.Fields(message.CommandArguments())
	var targetUser *tgbotapi.User
	var duration int
	var reason string
	var err error

	if message.ReplyToMessage != nil && message.ReplyToMessage.From != nil {
		// Reply-to-Message: /mute [Stunden] [Grund]
		targetUser = message.ReplyToMessage.From

		if len(args) > 0 {
			// Versuche ersten Arg als Dauer zu parsen
			duration, err = bot.ParseDuration(args[0])
			if err != nil {
				// Wenn nicht parsebar, ist es Teil des Grunds
				duration = b.GetConfig().Admin.DefaultMuteHours
				reason = strings.Join(args, " ")
			} else {
				// Dauer erfolgreich geparst, Rest ist Grund
				if len(args) > 1 {
					reason = strings.Join(args[1:], " ")
				}
			}
		} else {
			duration = b.GetConfig().Admin.DefaultMuteHours
		}
	} else {
		// Normal: /mute @user [Stunden] [Grund]
		if len(args) < 1 {
			return nil, 0, "", fmt.Errorf("Verwendung: /mute @username [Stunden] [Grund] oder als Antwort auf eine Nachricht")
		}

		targetUser, err = parseUserFromArgs(b, message.Chat.ID, args[0])
		if err != nil {
			return nil, 0, "", err
		}

		if len(args) > 1 {
			// Versuche zweiten Arg als Dauer zu parsen
			duration, err = bot.ParseDuration(args[1])
			if err != nil {
				// Wenn nicht parsebar, ist alles ab Arg 1 der Grund
				duration = b.GetConfig().Admin.DefaultMuteHours
				reason = strings.Join(args[1:], " ")
			} else {
				// Dauer erfolgreich geparst, Rest ist Grund
				if len(args) > 2 {
					reason = strings.Join(args[2:], " ")
				}
			}
		} else {
			duration = b.GetConfig().Admin.DefaultMuteHours
		}
	}

	if targetUser.ID == 0 {
		return nil, 0, "", fmt.Errorf("User-ID ist 0 in Reply-Nachricht")
	}
	if targetUser.IsBot {
		return nil, 0, "", fmt.Errorf("Kann keine Aktionen auf Bots ausführen")
	}

	return targetUser, duration, reason, nil
}

func (h *MuteHandler) unmuteUser(b *bot.Bot, chatID, userID int64) {
	isMuted, err := b.GetDB().IsUserMuted(userID, chatID)
	if err != nil || !isMuted {
		return
	}

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

	b.RestrictChatMember(chatID, userID, permissions)
	b.GetDB().RemoveMutedUser(userID, chatID)
}

func (h *DeleteHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	args := update.Message.CommandArguments()
	if args == "" {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Verwendung: /del [Anzahl]\nBeispiel: /del 10")
		return nil
	}

	count, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Ungültige Anzahl. Bitte gib eine Zahl ein.")
		return nil
	}

	if count < 1 || count > b.GetConfig().Admin.MaxDeleteMessages {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, fmt.Sprintf("Anzahl muss zwischen 1 und %d liegen.", b.GetConfig().Admin.MaxDeleteMessages))
		return nil
	}

	startMessageID := update.Message.MessageID
	deletedCount := 0

	// Erweiterte Löschstrategie: Versuche einen größeren Bereich um aktuelle Nachricht
	// Da Message IDs nicht sequenziell sind, probieren wir verschiedene Bereiche
	maxTries := count * 3 // Erweitere Suchbereich

	for i := 1; i <= maxTries && deletedCount < count; i++ {
		// Versuche Nachrichten in beide Richtungen
		candidates := []int{
			startMessageID - i, // Rückwärts
			startMessageID + i, // Vorwärts
		}

		for _, messageID := range candidates {
			if messageID > 0 && deletedCount < count {
				err := b.DeleteMessage(update.Message.Chat.ID, messageID)
				if err == nil {
					deletedCount++
				}
			}
		}

		time.Sleep(30 * time.Millisecond)
	}

	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)

	if deletedCount > 0 {
		successMsg := fmt.Sprintf(
			"%d Nachrichten geloescht\n\n"+
				"Admin: %s",
			deletedCount,
			bot.GetUserMention(update.Message.From),
		)

		_, _ = b.SendTemporaryMessageAndDeleteCommand(update.Message.Chat.ID, successMsg, update.Message.MessageID, 5)
	}

	return nil
}

func extractTargetUserAndReason(b *bot.Bot, message *tgbotapi.Message) (*tgbotapi.User, string, error) {
	var reason string

	// Priority 1: Reply to message
	if message.ReplyToMessage != nil && message.ReplyToMessage.From != nil {
		user := message.ReplyToMessage.From
		log.Printf("DEBUG: Reply-to-Message User: ID=%d, Username=%s, FirstName=%s",
			user.ID, user.UserName, user.FirstName)

		if user.ID == 0 {
			return nil, "", fmt.Errorf("User-ID ist 0 in Reply-Nachricht")
		}
		if user.IsBot {
			return nil, "", fmt.Errorf("Kann keine Aktionen auf Bots ausführen")
		}

		// Reason aus Command-Arguments extrahieren
		args := message.CommandArguments()
		if args != "" {
			reason = strings.TrimSpace(args)
		}

		return user, reason, nil
	}

	// Priority 2: Parse from command arguments
	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		return nil, "", fmt.Errorf("Verwendung: /%s @username [Grund] oder als Antwort auf eine Nachricht", message.Command())
	}

	user, err := parseUserFromArgs(b, message.Chat.ID, args[0])
	if err != nil {
		return nil, "", err
	}

	// Reason aus weiteren Arguments extrahieren
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	return user, reason, nil
}

func extractTargetUser(b *bot.Bot, message *tgbotapi.Message) (*tgbotapi.User, error) {
	user, _, err := extractTargetUserAndReason(b, message)
	return user, err
}

func parseUserFromArgs(b *bot.Bot, chatID int64, arg string) (*tgbotapi.User, error) {
	arg = strings.TrimSpace(arg)

	// Try to parse as user ID first
	if userID, err := strconv.ParseInt(arg, 10, 64); err == nil {
		return &tgbotapi.User{
			ID: userID,
		}, nil
	}

	// If it starts with @, try to resolve username
	if strings.HasPrefix(arg, "@") {
		username := strings.TrimPrefix(arg, "@")
		userID, err := resolveUsernameInChat(b, chatID, username)
		if err != nil {
			return nil, err
		}
		return &tgbotapi.User{
			ID:       userID,
			UserName: username,
		}, nil
	}

	return nil, fmt.Errorf("Ungültiges Format: %s (verwende @username oder User-ID)", arg)
}

// resolveUsernameInChat versucht einen Username über Chat Member API aufzulösen
func resolveUsernameInChat(b *bot.Bot, chatID int64, username string) (int64, error) {
	// Telegram Bot API kann Usernames nicht direkt auflösen
	// Wir müssen den User über eine Nachricht finden oder Chat-Member enumerieren

	// Versuche Chat Member API (funktioniert nur bei kleinen Gruppen)
	config := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	}

	admins, err := b.GetAPI().GetChatAdministrators(config)
	if err == nil {
		for _, admin := range admins {
			if admin.User.UserName == username {
				return admin.User.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("Username @%s nicht gefunden. Bei großen Gruppen verwende 'Auf Nachricht antworten' oder User-ID", username)
}

type UnmuteHandler struct{}

func NewUnmuteHandler() *UnmuteHandler {
	return &UnmuteHandler{}
}

func (h *UnmuteHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" {
		return nil
	}

	isAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Du hast keine Berechtigung für diesen Befehl.", 5)
		return nil
	}

	targetUser, err := extractTargetUser(b, update.Message)
	if err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, ""+err.Error(), 5)
		return nil
	}

	isMuted, err := b.GetDB().IsUserMuted(targetUser.ID, update.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to check mute status: %w", err)
	}

	if !isMuted {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Dieser User ist nicht gemutet.")
		return nil
	}

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

	if err := b.RestrictChatMember(update.Message.Chat.ID, targetUser.ID, permissions); err != nil {
		_, _ = b.SendTemporaryGroupMessage(update.Message.Chat.ID, "Fehler beim Entmuten des Users.")
		return fmt.Errorf("failed to unmute user: %w", err)
	}

	if err := b.GetDB().RemoveMutedUser(targetUser.ID, update.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to remove muted user from database: %w", err)
	}

	successMsg := fmt.Sprintf(
		"User entmutet\n\n"+
			"User: %s\n"+
			"Admin: %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	_, _ = b.SendTemporaryMessageAndDeleteCommand(update.Message.Chat.ID, successMsg, update.Message.MessageID, 5)
	return nil
}
