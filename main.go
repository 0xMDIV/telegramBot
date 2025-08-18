package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"telegramBot/config"
	"telegramBot/pkg/admin"
	"telegramBot/pkg/bot"
	"telegramBot/pkg/captcha"
	"telegramBot/pkg/handlers"
)

func main() {
	configPath := flag.String("config", "config/config.json", "Path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	botInstance, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	registerHandlers(botInstance)

	log.Println("Starting Telegram Security Bot...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := botInstance.Start(); err != nil {
			log.Fatalf("Bot error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down...")

	if err := botInstance.Stop(); err != nil {
		log.Printf("Error stopping bot: %v", err)
	}

	log.Println("Bot stopped")
}

func registerHandlers(b *bot.Bot) {
	b.RegisterHandler("new_member", captcha.NewHandler())
	b.RegisterHandler("callback", captcha.NewCallbackHandler())
	b.RegisterHandler("message", handlers.NewMessageHandler())

	b.RegisterHandler("ban", admin.NewBanHandler())
	b.RegisterHandler("kick", admin.NewKickHandler())
	b.RegisterHandler("mute", admin.NewMuteHandler())
	b.RegisterHandler("unmute", admin.NewUnmuteHandler())
	b.RegisterHandler("del", admin.NewDeleteHandler())
	b.RegisterHandler("help", admin.NewHelpHandler())
	b.RegisterHandler("permissions", admin.NewPermissionsHandler())
}
