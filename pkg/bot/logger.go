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

func GetUserIdentifier(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}
	return fmt.Sprintf("ID:%d", user.ID)
}
