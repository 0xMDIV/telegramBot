package admin

import (
	"fmt"
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
		b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	targetUser, err := extractTargetUser(update.Message)
	if err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ "+err.Error())
		return nil
	}

	if targetUser.ID == update.Message.From.ID {
		b.SendMessage(update.Message.Chat.ID, "❌ Du kannst dich nicht selbst bannen.")
		return nil
	}

	targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
	if err != nil {
		return fmt.Errorf("failed to check target admin status: %w", err)
	}

	if targetIsAdmin {
		b.SendMessage(update.Message.Chat.ID, "❌ Admins können nicht gebannt werden.")
		return nil
	}

	if err := b.BanChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ Fehler beim Bannen des Users.")
		return fmt.Errorf("failed to ban user: %w", err)
	}

	successMsg := fmt.Sprintf(
		"🔨 **User gebannt**\n\n"+
			"👤 **User:** %s\n"+
			"👮 **Admin:** %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	b.SendMessage(update.Message.Chat.ID, successMsg)
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
		b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	targetUser, err := extractTargetUser(update.Message)
	if err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ "+err.Error())
		return nil
	}

	if targetUser.ID == update.Message.From.ID {
		b.SendMessage(update.Message.Chat.ID, "❌ Du kannst dich nicht selbst kicken.")
		return nil
	}

	targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
	if err != nil {
		return fmt.Errorf("failed to check target admin status: %w", err)
	}

	if targetIsAdmin {
		b.SendMessage(update.Message.Chat.ID, "❌ Admins können nicht gekickt werden.")
		return nil
	}

	if err := b.KickChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ Fehler beim Kicken des Users.")
		return fmt.Errorf("failed to kick user: %w", err)
	}

	if err := b.UnbanChatMember(update.Message.Chat.ID, targetUser.ID); err != nil {
		return fmt.Errorf("failed to unban user after kick: %w", err)
	}

	successMsg := fmt.Sprintf(
		"👢 **User gekickt**\n\n"+
			"👤 **User:** %s\n"+
			"👮 **Admin:** %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	b.SendMessage(update.Message.Chat.ID, successMsg)
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
		b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	args := strings.Fields(update.Message.CommandArguments())
	if len(args) < 1 {
		b.SendMessage(update.Message.Chat.ID, "❌ Verwendung: /mute @username [Stunden]\nBeispiel: /mute @user 2")
		return nil
	}

	var targetUser *tgbotapi.User
	var duration int

	if update.Message.ReplyToMessage != nil {
		targetUser = update.Message.ReplyToMessage.From
		if len(args) > 0 {
			duration, err = bot.ParseDuration(args[0])
			if err != nil {
				duration = b.GetConfig().Admin.DefaultMuteHours
			}
		} else {
			duration = b.GetConfig().Admin.DefaultMuteHours
		}
	} else {
		targetUser, err = parseUserFromArgs(args[0])
		if err != nil {
			b.SendMessage(update.Message.Chat.ID, "❌ "+err.Error())
			return nil
		}

		if len(args) > 1 {
			duration, err = bot.ParseDuration(args[1])
			if err != nil {
				b.SendMessage(update.Message.Chat.ID, "❌ "+err.Error())
				return nil
			}
		} else {
			duration = b.GetConfig().Admin.DefaultMuteHours
		}
	}

	if targetUser.ID == update.Message.From.ID {
		b.SendMessage(update.Message.Chat.ID, "❌ Du kannst dich nicht selbst muten.")
		return nil
	}

	targetIsAdmin, err := b.IsUserAdmin(update.Message.Chat.ID, targetUser.ID)
	if err != nil {
		return fmt.Errorf("failed to check target admin status: %w", err)
	}

	if targetIsAdmin {
		b.SendMessage(update.Message.Chat.ID, "❌ Admins können nicht gemutet werden.")
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
		CanAddWebPagePreviates: false,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
	}

	if err := b.RestrictChatMember(update.Message.Chat.ID, targetUser.ID, permissions); err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ Fehler beim Muten des Users.")
		return fmt.Errorf("failed to mute user: %w", err)
	}

	successMsg := fmt.Sprintf(
		"🔇 **User gemutet**\n\n"+
			"👤 **User:** %s\n"+
			"⏰ **Dauer:** %d Stunden\n"+
			"📅 **Bis:** %s\n"+
			"👮 **Admin:** %s",
		bot.FormatUserName(targetUser),
		duration,
		muteUntil.Format("02.01.2006 15:04"),
		bot.GetUserMention(update.Message.From),
	)

	b.SendMessage(update.Message.Chat.ID, successMsg)

	go func() {
		time.Sleep(time.Duration(duration) * time.Hour)
		h.unmuteUser(b, update.Message.Chat.ID, targetUser.ID)
	}()

	return nil
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
		b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	args := update.Message.CommandArguments()
	if args == "" {
		b.SendMessage(update.Message.Chat.ID, "❌ Verwendung: /del [Anzahl]\nBeispiel: /del 10")
		return nil
	}

	count, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ Ungültige Anzahl. Bitte gib eine Zahl ein.")
		return nil
	}

	if count < 1 || count > b.GetConfig().Admin.MaxDeleteMessages {
		b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("❌ Anzahl muss zwischen 1 und %d liegen.", b.GetConfig().Admin.MaxDeleteMessages))
		return nil
	}

	startMessageID := update.Message.MessageID
	deletedCount := 0

	for i := 0; i < count; i++ {
		messageID := startMessageID - i - 1
		if messageID <= 0 {
			break
		}

		err := b.DeleteMessage(update.Message.Chat.ID, messageID)
		if err == nil {
			deletedCount++
		}
		
		time.Sleep(50 * time.Millisecond)
	}

	b.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID)

	if deletedCount > 0 {
		successMsg := fmt.Sprintf(
			"🗑️ **%d Nachrichten gelöscht**\n\n"+
				"👮 **Admin:** %s",
			deletedCount,
			bot.GetUserMention(update.Message.From),
		)

		msg, err := b.SendMessage(update.Message.Chat.ID, successMsg)
		if err == nil {
			go func() {
				time.Sleep(5 * time.Second)
				b.DeleteMessage(update.Message.Chat.ID, msg.MessageID)
			}()
		}
	}

	return nil
}

func extractTargetUser(message *tgbotapi.Message) (*tgbotapi.User, error) {
	if message.ReplyToMessage != nil {
		return message.ReplyToMessage.From, nil
	}

	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		return nil, fmt.Errorf("Verwendung: /%s @username oder als Antwort auf eine Nachricht", message.Command())
	}

	return parseUserFromArgs(args[0])
}

func parseUserFromArgs(arg string) (*tgbotapi.User, error) {
	arg = strings.TrimSpace(arg)
	
	if strings.HasPrefix(arg, "@") {
		username := strings.TrimPrefix(arg, "@")
		return &tgbotapi.User{
			UserName: username,
		}, nil
	}

	userID, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Ungültiger Username oder User-ID: %s", arg)
	}

	return &tgbotapi.User{
		ID: userID,
	}, nil
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
		b.SendMessage(update.Message.Chat.ID, "❌ Du hast keine Berechtigung für diesen Befehl.")
		return nil
	}

	targetUser, err := extractTargetUser(update.Message)
	if err != nil {
		b.SendMessage(update.Message.Chat.ID, "❌ "+err.Error())
		return nil
	}

	isMuted, err := b.GetDB().IsUserMuted(targetUser.ID, update.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to check mute status: %w", err)
	}

	if !isMuted {
		b.SendMessage(update.Message.Chat.ID, "❌ Dieser User ist nicht gemutet.")
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
		b.SendMessage(update.Message.Chat.ID, "❌ Fehler beim Entmuten des Users.")
		return fmt.Errorf("failed to unmute user: %w", err)
	}

	if err := b.GetDB().RemoveMutedUser(targetUser.ID, update.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to remove muted user from database: %w", err)
	}

	successMsg := fmt.Sprintf(
		"🔊 **User entmutet**\n\n"+
			"👤 **User:** %s\n"+
			"👮 **Admin:** %s",
		bot.FormatUserName(targetUser),
		bot.GetUserMention(update.Message.From),
	)

	b.SendMessage(update.Message.Chat.ID, successMsg)
	return nil
}