# Telegram Security Bot

Ein modularer Telegram-Bot für Gruppensicherheit mit erweiterten Moderationsfunktionen.

## 🚀 Features

- **Captcha-System**: Neue Mitglieder müssen ein mathematisches Captcha per DM lösen
- **Admin-Commands**: Umfangreiche Moderationsbefehle für Admins
- **Multi-Gruppen-Support**: Ein Bot kann mehrere Gruppen gleichzeitig verwalten
- **Automatisches Muting**: Gemutete Benutzer können keine Nachrichten senden
- **Modularer Aufbau**: Einfach erweiterbar durch Handler-System

## 📋 Verfügbare Commands

### Admin-Commands (nur für Gruppen-Admins)

- `/ban @user` - Bannt einen User permanent aus der Gruppe
- `/kick @user` - Kickt einen User aus der Gruppe (kann später wieder beitreten)
- `/mute @user [Stunden]` - Mutet einen User für X Stunden (Standard: 1 Stunde)
- `/unmute @user` - Entfernt das Mute von einem User
- `/del [Anzahl]` - Löscht die letzten X Nachrichten (max. 100)

Alle Commands funktionieren auch als Antwort auf Nachrichten:
```
/ban
/kick
/mute 2
```

## 🛠️ Installation

### Voraussetzungen

- Go 1.21 oder höher
- SQLite3 (für CGO)
- Ein Telegram Bot Token (von @BotFather)

### 1. Repository klonen

```bash
git clone <repository-url>
cd telegramBot
```

### 2. Dependencies installieren

```bash
go mod download
```

### 3. Konfiguration

Kopiere die `config/config.json` und trage deinen Bot-Token ein:

```json
{
  "bot_token": "DEIN_BOT_TOKEN_HIER",
  "debug": false,
  "captcha": {
    "timeout_minutes": 5,
    "max_attempts": 3,
    "welcome_message": "Willkommen in der Gruppe! 🎉"
  },
  "admin": {
    "default_mute_hours": 1,
    "max_delete_messages": 100
  },
  "database": {
    "file_path": "bot_data.db"
  }
}
```

### 4. Bot-Setup

1. Erstelle einen Bot bei [@BotFather](https://t.me/botfather)
2. Erhalte den Bot-Token
3. Füge den Bot zu deiner Gruppe hinzu
4. Mache den Bot zum Admin mit folgenden Rechten:
   - Delete messages
   - Ban users
   - Restrict members
   - Add new admins (optional)

## 🏃‍♂️ Bot starten

### Entwicklung

```bash
go run .
```

### Mit custom Config

```bash
go run . -config=/path/to/config.json
```

### Production Build

#### Automatisch (alle Plattformen)

```bash
# Linux/macOS
./build.sh

# Windows
build.bat

# Oder mit Make
make build
```

#### Manuell

```bash
# Windows
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/telegram-security-bot.exe .

# Linux
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/telegram-security-bot .
```

## 🔧 Deployment

### Linux Server

1. Binary hochladen und ausführbar machen:
```bash
chmod +x telegram-security-bot-linux-amd64
```

2. Systemd Service erstellen (`/etc/systemd/system/telegram-bot.service`):
```ini
[Unit]
Description=Telegram Security Bot
After=network.target

[Service]
Type=simple
User=telegram-bot
ExecStart=/path/to/telegram-security-bot-linux-amd64 -config=/path/to/config.json
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. Service starten:
```bash
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

### Windows Server

1. Als Windows Service installieren (mit NSSM oder sc.exe)
2. Oder einfach als normale Anwendung starten:
```cmd
telegram-security-bot-windows-amd64.exe -config=config\config.json
```

## 🔒 Sicherheitsfeatures

### Captcha-System

- Neue Mitglieder werden automatisch stummgeschaltet
- Captcha wird per DM gesendet (mathematische Aufgabe)
- 3 Versuche, 5 Minuten Zeit
- Bei Fehlern: automatischer Kick aus der Gruppe
- Nach erfolgreichem Lösen: Willkommensnachricht (auto-delete nach 10s)

### Mute-System

- Gemutete User können keine Nachrichten senden
- Automatisches Löschen von Nachrichten gemuteter User
- Automatisches Entmuten nach Ablauf der Zeit
- Persistent in der Datenbank gespeichert

## 🗄️ Datenbank

Der Bot verwendet SQLite für die Datenpersistierung:

- `pending_users` - Captcha-Daten
- `muted_users` - Mute-Status und Dauer
- `group_settings` - Gruppenspezifische Einstellungen

Die Datenbank wird automatisch beim ersten Start erstellt.

## 🏗️ Architektur

```
telegramBot/
├── config/              # Konfigurationsdateien
├── pkg/
│   ├── bot/            # Bot-Core und Handler-Interface
│   ├── database/       # Datenbankoperationen
│   ├── captcha/        # Captcha-System
│   ├── admin/          # Admin-Commands
│   └── handlers/       # Message-Handler
├── cmd/bot/            # Alternative Main-Implementierung
└── main.go             # Hauptanwendung
```

### Handler-System

Jeder Handler implementiert das `Handler` Interface:

```go
type Handler interface {
    Handle(bot *Bot, update tgbotapi.Update) error
}
```

Neue Features können einfach durch neue Handler hinzugefügt werden.

## 🔌 Erweiterungen

### Neuen Command hinzufügen

1. Handler erstellen:
```go
type MyHandler struct{}

func (h *MyHandler) Handle(b *bot.Bot, update tgbotapi.Update) error {
    // Command-Logik hier
    return nil
}
```

2. Handler registrieren:
```go
b.RegisterHandler("mycommand", &MyHandler{})
```

### Neue Konfigurationsoptionen

1. Config-Struct erweitern (`config/config.go`)
2. Default-Werte in `config.json` hinzufügen
3. In Handlers verwenden: `b.GetConfig()`

## 📊 Logging

Der Bot loggt wichtige Ereignisse:
- Startup/Shutdown
- Command-Ausführungen
- Errors und Warnings
- Captcha-Ereignisse

## 🚨 Troubleshooting

### Bot reagiert nicht

1. Prüfe Bot-Token in config.json
2. Stelle sicher, dass der Bot Admin-Rechte hat
3. Prüfe Logs auf Fehler

### Captcha funktioniert nicht

1. Bot muss Nachrichten an User senden können
2. User muss zuerst eine Unterhaltung mit dem Bot starten
3. Prüfe `CanSendMessages` Berechtigung

### Befehle funktionieren nicht

1. Nur Gruppen-Admins können Admin-Commands nutzen
2. Bot braucht entsprechende Admin-Rechte
3. Commands sind case-sensitive

## 🤝 Contributing

1. Fork das Repository
2. Erstelle einen Feature-Branch
3. Committe deine Änderungen
4. Erstelle einen Pull Request

## 📄 Lizenz

MIT License - siehe LICENSE Datei für Details.

## 🔗 Links

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Go Telegram Bot Library](https://github.com/go-telegram-bot-api/telegram-bot-api)