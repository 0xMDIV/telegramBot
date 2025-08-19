# Telegram Security Bot

Ein modularer Telegram-Bot für Gruppensicherheit mit erweiterten Moderationsfunktionen.

## 🚀 Features

- **Gruppen-basiertes Captcha-System**: Neue Mitglieder lösen Captcha direkt in der Gruppe
- **Admin-Commands**: Umfangreiche Moderationsbefehle für Admins
- **Konfigurationsverwaltung**: Alle Einstellungen per DM-Chat anpassbar
- **Admin-Management**: Dynamisches Hinzufügen/Entfernen von Bot-Admins
- **Multi-Gruppen-Support**: Ein Bot kann mehrere Gruppen gleichzeitig verwalten
- **Automatisches Muting**: Gemutete Benutzer können keine Nachrichten senden
- **Umfassendes Logging**: Alle Events, Commands und Nachrichten werden geloggt
- **Modularer Aufbau**: Einfach erweiterbar durch Handler-System

## 📋 Verfügbare Commands

### Admin-Commands (für Gruppen-Admins & Bot-Admins)

#### Moderation
- `/ban @user [Grund]` - Bannt einen User permanent aus der Gruppe
- `/kick @user [Grund]` - Kickt einen User aus der Gruppe (kann später wieder beitreten)
- `/mute @user [Stunden] [Grund]` - Mutet einen User für X Stunden (Standard: 1 Stunde)
- `/unmute @user` - Entfernt das Mute von einem User
- `/del [Anzahl]` - Löscht die letzten X Nachrichten (max. 100)

#### Admin-Management
- `/add_admin @user` - Fügt einen User als Bot-Admin hinzu
- `/add_admin 123456789` - Fügt einen User per ID als Bot-Admin hinzu
- `/del_admin @user` - Entfernt einen User als Bot-Admin
- `/del_admin 123456789` - Entfernt einen User per ID als Bot-Admin

#### Hilfsbefehle
- `/help` - Zeigt alle verfügbaren Commands
- `/permissions` - Zeigt aktuelle Berechtigungen

### Konfiguration (nur für Bot-Admins per DM)

- `/config` - Zeigt alle konfigurierbaren Einstellungen mit aktuellen Werten
- `/config <schlüssel> <wert>` - Ändert eine Konfiguration

#### Verfügbare Konfigurationsschlüssel:
- `timeout_minutes` - Zeitlimit für Captcha (1-60 Min)
- `max_attempts` - Maximale Captcha-Versuche (1-10)
- `welcome_message` - Willkommensnachricht für neue User
- `message_delete_delay_minutes` - Löschzeit für Willkommensnachrichten (1-60 Min)
- `success_message_delete_delay_minutes` - Löschzeit für Erfolgsnachrichten (1-60 Min)
- `default_mute_hours` - Standard Mute-Dauer (1-168 Std)
- `max_delete_messages` - Max löschbare Nachrichten (1-1000)

#### Konfigurationsbeispiele:
```
/config timeout_minutes 10
/config welcome_message "Hallo! Willkommen!"
/config success_message_delete_delay_minutes 2
```

### Command-Verwendung

Alle Commands funktionieren auch als Antwort auf Nachrichten:
```
/ban              # Bannt den User der ursprünglichen Nachricht
/kick Spam        # Kickt mit Grund "Spam"
/mute 2 Störend   # Mutet für 2 Stunden mit Grund "Störend"
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
    "welcome_message": "Willkommen in der Gruppe! 🎉",
    "message_delete_delay_minutes": 5,
    "success_message_delete_delay_minutes": 3
  },
  "admin": {
    "default_mute_hours": 1,
    "max_delete_messages": 100,
    "admin_user_ids": []
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

### 5. Erste Admin-Konfiguration

Füge deine User-ID zur `admin_user_ids` Liste in der config.json hinzu oder verwende `/add_admin` nachdem der Bot läuft.

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

### Gruppen-Captcha-System

**Neuer Ablauf:**
1. **User joint** → Bot sendet Willkommensnachricht mit Captcha direkt in der Gruppe
2. **User antwortet** mit der richtigen Zahl in die Gruppe
3. **Bei Erfolg**: User bekommt volle Berechtigung, Nachrichten werden nach konfigurierbarer Zeit gelöscht
4. **Bei Fehlschlag**: User wird nach zu vielen Versuchen oder Timeout automatisch gekickt

**Eigenschaften:**
- Captcha erfolgt **direkt in der Gruppe** (keine DM-Probleme mehr)
- Mathematische Aufgaben (z.B. "5+3 = ?")
- Konfigurierbare Zeitlimits und Versuche
- Separate Löschzeiten für verschiedene Nachrichtentypen
- Umfassendes Logging aller Captcha-Events

### Mute-System

- Gemutete User können keine Nachrichten senden
- Automatisches Löschen von Nachrichten gemuteter User
- Automatisches Entmuten nach Ablauf der Zeit
- Persistent in der Datenbank gespeichert

### Admin-System

**Zwei Admin-Ebenen:**
1. **Gruppen-Admins**: Haben automatisch alle Bot-Rechte in ihrer Gruppe
2. **Bot-Admins**: Sind in der Config gespeichert, haben globale Rechte

**Admin-Management:**
- Dynamisches Hinzufügen/Entfernen von Bot-Admins über Commands
- Beide Admin-Typen können `/add_admin` und `/del_admin` verwenden
- Config-Änderungen nur für Bot-Admins per DM

## 📊 Logging-System

Der Bot erstellt zwei separate Log-Dateien:

### commands.log
```
[2025-08-19 23:31:00] Chat: -123456 | User: 123456 (@admin) | Command: ban @spammer | Result: SUCCESS
```

### events.log
```
[2025-08-19 23:30:15] USER_JOINED | Chat: -123456 | User: 84937883 (@username) | Details: User joined the group
[2025-08-19 23:30:45] MESSAGE | Chat: -123456 | User: 84937883 (@username) | Text: 8
[2025-08-19 23:30:50] CAPTCHA_SUCCESS | Chat: -123456 | User: 84937883 (@username) | Details: Captcha solved successfully after 1 attempts
[2025-08-19 23:32:00] CAPTCHA_FAIL | Chat: -123456 | User: 84937883 (@username) | Details: Too many wrong attempts
[2025-08-19 23:32:01] USER_KICKED | Chat: -123456 | User: 84937883 (@username) | Details: Captcha failed - too many attempts
```

**Geloggte Events:**
- Alle User-Joins und Leaves
- Alle Nachrichten (außer Commands)
- Alle Captcha-Erfolge und Fehlschläge
- Alle Kicks und deren Gründe
- Alle Admin-Commands und deren Ergebnisse

## 🗄️ Datenbank

Der Bot verwendet SQLite für die Datenpersistierung:

- `pending_users` - Captcha-Daten und Versuche
- `muted_users` - Mute-Status und Dauer
- `group_settings` - Gruppenspezifische Einstellungen
- `welcome_messages` - Tracking von Willkommensnachrichten für Löschung

Die Datenbank wird automatisch beim ersten Start erstellt.

## 🏗️ Architektur

```
telegramBot/
├── config/              # Konfigurationsdateien
├── pkg/
│   ├── bot/            # Bot-Core, Handler-Interface und Logging
│   ├── database/       # Datenbankoperationen
│   ├── captcha/        # Captcha-System (Gruppen + Message Handler)
│   ├── admin/          # Admin-Commands + Config-Management
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

**Registrierte Handler:**
- `new_member` - Captcha für neue User
- `captcha_message` - Captcha-Antworten verarbeiten (vor normalem Message-Handler)
- `message` - Normale Nachrichten
- `callback` - Callback-Queries (Legacy)
- Admin-Commands: `ban`, `kick`, `mute`, `unmute`, `del`, `help`, `permissions`
- Admin-Management: `add_admin`, `del_admin`, `config`

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
3. Config-Handler erweitern (`pkg/admin/config.go`)
4. In Handlers verwenden: `b.GetConfig()`

### Neues Logging hinzufügen

```go
// Event-Logging
b.GetEventLogger().LogEvent("CUSTOM_EVENT", chatID, userID, username, "Details")

// Command-Logging (automatisch für alle Commands)
// Wird automatisch von der handleUpdate-Funktion gemacht
```

## 🚨 Troubleshooting

### Bot reagiert nicht

1. Prüfe Bot-Token in config.json
2. Stelle sicher, dass der Bot Admin-Rechte hat
3. Prüfe Logs (`commands.log`, `events.log`) auf Fehler
4. Prüfe Konsolen-Output

### Captcha funktioniert nicht

1. **Neues System**: Captcha läuft jetzt direkt in der Gruppe
2. User müssen Nachrichten senden können (wird automatisch erlaubt)
3. Prüfe `message_delete_delay_minutes` Config
4. Prüfe Events-Log für Captcha-Events

### Config-Befehle funktionieren nicht

1. `/config` funktioniert nur per **DM** an den Bot
2. Nur **Bot-Admins** (in config.json) können Config ändern
3. Gruppen-Admins haben nur normale Admin-Commands
4. User-ID muss in `admin_user_ids` Array stehen

### Admin-Commands funktionieren nicht

1. **Gruppen-Admins** haben automatisch alle Bot-Admin-Rechte
2. **Bot-Admins** sind global berechtigt
3. Bot braucht entsprechende Admin-Rechte in der Gruppe
4. Commands sind case-sensitive

### Logging-Probleme

1. Prüfe Schreibrechte im Bot-Verzeichnis
2. Log-Dateien: `commands.log` und `events.log`
3. Bei Fehlern: Konsolen-Output prüfen

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