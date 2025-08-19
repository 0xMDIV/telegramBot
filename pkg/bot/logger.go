package bot

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandLogger struct {
	logFile *os.File
}

type EventLogger struct {
	logFile *os.File
}

func NewCommandLogger(filepath string) (*CommandLogger, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &CommandLogger{
		logFile: file,
	}, nil
}

func (cl *CommandLogger) LogCommand(chatID int64, userID int64, username, command, args, result string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := fmt.Sprintf("[%s] Chat: %d | User: %d (%s) | Command: %s %s | Result: %s\n",
		timestamp, chatID, userID, username, command, args, result)

	if _, err := cl.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to command log: %v", err)
	} else {
		cl.logFile.Sync()
	}
}

func (cl *CommandLogger) Close() error {
	if cl.logFile != nil {
		return cl.logFile.Close()
	}
	return nil
}

func NewEventLogger(filepath string) (*EventLogger, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log file: %w", err)
	}

	return &EventLogger{
		logFile: file,
	}, nil
}

func (el *EventLogger) LogEvent(eventType string, chatID int64, userID int64, username string, details string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := fmt.Sprintf("[%s] %s | Chat: %d | User: %d (%s) | Details: %s\n",
		timestamp, eventType, chatID, userID, username, details)

	if _, err := el.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to event log: %v", err)
	} else {
		el.logFile.Sync()
	}
}

func (el *EventLogger) LogMessage(chatID int64, userID int64, username string, messageText string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := fmt.Sprintf("[%s] MESSAGE | Chat: %d | User: %d (%s) | Text: %s\n",
		timestamp, chatID, userID, username, messageText)

	if _, err := el.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to event log: %v", err)
	} else {
		el.logFile.Sync()
	}
}

func (el *EventLogger) LogJoin(chatID int64, userID int64, username string) {
	el.LogEvent("USER_JOINED", chatID, userID, username, "User joined the group")
}

func (el *EventLogger) LogLeave(chatID int64, userID int64, username string) {
	el.LogEvent("USER_LEFT", chatID, userID, username, "User left the group")
}

func (el *EventLogger) LogCaptchaSuccess(chatID int64, userID int64, username string, attempts int) {
	details := fmt.Sprintf("Captcha solved successfully after %d attempts", attempts)
	el.LogEvent("CAPTCHA_SUCCESS", chatID, userID, username, details)
}

func (el *EventLogger) LogCaptchaFail(chatID int64, userID int64, username string, reason string) {
	el.LogEvent("CAPTCHA_FAIL", chatID, userID, username, reason)
}

func (el *EventLogger) LogKick(chatID int64, userID int64, username string, reason string) {
	el.LogEvent("USER_KICKED", chatID, userID, username, reason)
}

func (el *EventLogger) Close() error {
	if el.logFile != nil {
		return el.logFile.Close()
	}
	return nil
}

func GetUserIdentifier(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}
	return fmt.Sprintf("ID:%d", user.ID)
}
