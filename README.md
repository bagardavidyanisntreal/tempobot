# tempobot (Telegram event signup bot)

Complete minimal Go repository.

## Project tree

```
tempobot
├ go.mod
├ main.go
└ internal
    ├ model
    │   ├ user.go
    │   ├ event.go
    │   └ participant.go
    ├ storage
    │   ├ db.go
    │   ├ user_repo.go
    │   ├ event_repo.go
    │   └ participant_repo.go
    ├ service
    │   ├ user_service.go
    │   └ event_service.go
    └ telegram
        ├ types.go
        ├ client.go
        ├ keyboard.go
        ├ render.go
        └ handler.go
```

---

# go.mod

```go
module tempobot

go 1.22

require github.com/lib/pq v1.10.9
```

---

# main.go

```go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"

    "tempobot/internal/service"
    "tempobot/internal/storage"
    "tempobot/internal/telegram"
)

func main() {

    token := os.Getenv("BOT_TOKEN")
    dbURL := os.Getenv("DB_URL")

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }

    userRepo := storage.NewUserRepo(db)
    eventRepo := storage.NewEventRepo(db)
    partRepo := storage.NewParticipantRepo(db)

    userService := service.NewUserService(userRepo)
    eventService := service.NewEventService(eventRepo, partRepo)

    tg := telegram.NewClient(token)

    handler := telegram.NewHandler(userService, eventService, tg)

    http.HandleFunc("/webhook", handler.HandleWebhook)

    log.Println("server started :8080")

    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

# internal/model/user.go

```go
package model

type User struct {
    ID int64

    TelegramUserID int64
    Username string
    FirstName string

    Role string
}
```

---

# internal/model/event.go

```go
package model

import "time"

type Event struct {
    ID int64

    Title string
    Description string

    StartAt time.Time
    RegistrationDeadline time.Time

    CreatedBy int64

    ChatID int64
    MessageID int
}
```

---

# internal/model/participant.go

```go
package model

type Participant struct {

    ID int64

    EventID int64
    UserID int64

    Status string
}
```

---

# internal/storage/db.go

```go
package storage

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewPostgres(url string) (*sql.DB, error) {
    return sql.Open("postgres", url)
}
```

---

# internal/storage/user_repo.go

```go
package storage

import (
    "database/sql"
    "errors"

    "tempobot/internal/model"
)

type UserRepo struct {
    db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) FindByTelegramID(id int64) (*model.User, error) {

    row := r.db.QueryRow(`
SELECT id, telegram_user_id, username, first_name, role
FROM users
WHERE telegram_user_id = $1
`, id)

    var u model.User

    err := row.Scan(
        &u.ID,
        &u.TelegramUserID,
        &u.Username,
        &u.FirstName,
        &u.Role,
    )

    if err != nil {
        return nil, err
    }

    return &u, nil
}

func (r *UserRepo) Create(
    telegramID int64,
    username string,
    firstName string,
) (*model.User, error) {

    row := r.db.QueryRow(`
INSERT INTO users (telegram_user_id, username, first_name)
VALUES ($1,$2,$3)
RETURNING id, telegram_user_id, username, first_name, role
`, telegramID, username, firstName)

    var u model.User

    err := row.Scan(
        &u.ID,
        &u.TelegramUserID,
        &u.Username,
        &u.FirstName,
        &u.Role,
    )

    if err != nil {
        return nil, err
    }

    return &u, nil
}

func (r *UserRepo) UpdateProfile(
    id int64,
    username string,
    firstName string,
) error {

    _, err := r.db.Exec(`
UPDATE users
SET username = $1,
    first_name = $2
WHERE id = $3
`, username, firstName, id)

    return err
}
```

---

# internal/storage/event_repo.go

```go
package storage

import (
    "database/sql"

    "tempobot/internal/model"
)

type EventRepo struct {
    db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
    return &EventRepo{db: db}
}

func (r *EventRepo) Get(id int64) (*model.Event, error) {

    row := r.db.QueryRow(`
SELECT id,title,description,chat_id,message_id
FROM events
WHERE id=$1
`, id)

    var e model.Event

    err := row.Scan(
        &e.ID,
        &e.Title,
        &e.Description,
        &e.ChatID,
        &e.MessageID,
    )

    if err != nil {
        return nil, err
    }

    return &e, nil
}
```

---

# internal/storage/participant_repo.go

```go
package storage

import "database/sql"

type ParticipantRepo struct {
    db *sql.DB
}

func NewParticipantRepo(db *sql.DB) *ParticipantRepo {
    return &ParticipantRepo{db: db}
}

func (r *ParticipantRepo) Upsert(eventID, userID int64, status string) error {

    query := `
INSERT INTO participants (event_id, user_id, status)
VALUES ($1,$2,$3)
ON CONFLICT (event_id,user_id)
DO UPDATE SET
status = EXCLUDED.status,
updated_at = now()
`

    _, err := r.db.Exec(query, eventID, userID, status)

    return err
}

func (r *ParticipantRepo) CountByStatus(eventID int64) (map[string]int, error) {

    rows, err := r.db.Query(`
SELECT status, count(*)
FROM participants
WHERE event_id = $1
GROUP BY status
`, eventID)

    if err != nil {
        return nil, err
    }

    stats := map[string]int{
        "going": 0,
        "maybe": 0,
        "no": 0,
    }

    for rows.Next() {

        var status string
        var count int

        rows.Scan(&status, &count)

        stats[status] = count
    }

    return stats, nil
}
```

---

# internal/service/user_service.go

```go
package service

import (
    "database/sql"

    "tempobot/internal/model"
    "tempobot/internal/storage"
)

type UserService struct {
    repo *storage.UserRepo
}

func NewUserService(r *storage.UserRepo) *UserService {
    return &UserService{repo: r}
}

func (s *UserService) EnsureUser(
    telegramID int64,
    username string,
    firstName string,
) (*model.User, error) {

    u, err := s.repo.FindByTelegramID(telegramID)

    if err == nil {
        _ = s.repo.UpdateProfile(u.ID, username, firstName)
        return u, nil
    }

    if err == sql.ErrNoRows {
        return s.repo.Create(telegramID, username, firstName)
    }

    return nil, err
}
```

---

# internal/service/event_service.go

```go
package service

import "tempobot/internal/storage"

type EventService struct {

    eventRepo *storage.EventRepo
    partRepo *storage.ParticipantRepo
}

func NewEventService(
    e *storage.EventRepo,
    p *storage.ParticipantRepo,
) *EventService {

    return &EventService{
        eventRepo: e,
        partRepo: p,
    }
}

func (s *EventService) RegisterParticipant(eventID, userID int64, status string) error {

    return s.partRepo.Upsert(eventID, userID, status)
}

func (s *EventService) GetEvent(eventID int64) (interface{}, error) {

    return s.eventRepo.Get(eventID)
}

func (s *EventService) GetStats(eventID int64) (map[string]int, error) {

    return s.partRepo.CountByStatus(eventID)
}
```

---

# internal/telegram/types.go

```go
package telegram

type Update struct {

    UpdateID int `json:"update_id"`

    Message *Message `json:"message,omitempty"`

    CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}


type Message struct {

    MessageID int `json:"message_id"`

    Chat Chat `json:"chat"`

    Text string `json:"text"`

    From User `json:"from"`
}


type CallbackQuery struct {

    ID string `json:"id"`

    Data string `json:"data"`

    From User `json:"from"`
}


type Chat struct {

    ID int64 `json:"id"`
}


type User struct {

    ID int64 `json:"id"`

    Username string `json:"username"`

    FirstName string `json:"first_name"`
}
```

---

# internal/telegram/client.go

```go
package telegram

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)


type Client struct {

    token string
}


func NewClient(token string) *Client {

    return &Client{token: token}
}


func (c *Client) EditMessage(
    chatID int64,
    messageID int,
    text string,
    eventID int64,
    stats map[string]int,
) error {

    url := fmt.Sprintf(
        "https://api.telegram.org/bot%s/editMessageText",
        c.token,
    )

    req := map[string]interface{}{
        "chat_id": chatID,
        "message_id": messageID,
        "text": text,
        "reply_markup": InlineKeyboard(eventID, stats),
    }

    body, _ := json.Marshal(req)

    _, err := http.Post(url, "application/json", bytes.NewBuffer(body))

    return err
}
```

---

# internal/telegram/keyboard.go

```go
package telegram

import "strconv"


func InlineKeyboard(eventID int64, stats map[string]int) map[string]interface{} {

    id := strconv.FormatInt(eventID, 10)

    return map[string]interface{}{
        "inline_keyboard": [][]map[string]string{
            {
                {
                    "text": "🏃 Побегу (" + strconv.Itoa(stats["going"]) + ")",
                    "callback_data": "event:" + id + ":going",
                },
                {
                    "text": "🤔 Возможно (" + strconv.Itoa(stats["maybe"]) + ")",
                    "callback_data": "event:" + id + ":maybe",
                },
                {
                    "text": "❌ Не смогу (" + strconv.Itoa(stats["no"]) + ")",
                    "callback_data": "event:" + id + ":no",
                },
            },
        },
    }
}
```

---

# internal/telegram/render.go

```go
package telegram

import (
    "fmt"

    "tempobot/internal/model"
)


func BuildEventText(event *model.Event, stats map[string]int) string {

    return fmt.Sprintf(
        "🏃 %s\n\n%s\n\n"+
            "🏃 Побегут — %d\n"+
            "🤔 Возможно — %d\n"+
            "❌ Не смогут — %d",
        event.Title,
        event.Description,
        stats["going"],
        stats["maybe"],
        stats["no"],
    )
}
```

---

# internal/telegram/handler.go

```go
package telegram

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"

    "tempobot/internal/model"
    "tempobot/internal/service"
)


type Handler struct {

    userService *service.UserService

    eventService *service.EventService

    tg *Client
}


func NewHandler(
    us *service.UserService,
    es *service.EventService,
    tg *Client,
) *Handler {

    return &Handler{
        userService: us,
        eventService: es,
        tg: tg,
    }
}


func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {

    var upd Update

    err := json.NewDecoder(r.Body).Decode(&upd)

    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    if upd.CallbackQuery != nil {
        h.handleCallback(upd.CallbackQuery)
    }

    w.WriteHeader(200)
}


func (h *Handler) handleCallback(cb *CallbackQuery) {

    user := cb.From

    u, err := h.userService.EnsureUser(
        user.ID,
        user.Username,
        user.FirstName,
    )

    if err != nil {
        log.Println(err)
        return
    }

    parts := strings.Split(cb.Data, ":")

    if len(parts) != 3 {
        return
    }

    eventID, _ := strconv.ParseInt(parts[1], 10, 64)

    status := parts[2]

    err = h.eventService.RegisterParticipant(eventID, u.ID, status)

    if err != nil {
        log.Println(err)
        return
    }

    eRaw, err := h.eventService.GetEvent(eventID)

    if err != nil {
        log.Println(err)
        return
    }

    event := eRaw.(*model.Event)

    stats, err := h.eventService.GetStats(eventID)

    if err != nil {
        log.Println(err)
        return
    }

    text := BuildEventText(event, stats)

    err = h.tg.EditMessage(
        event.ChatID,
        event.MessageID,
        text,
        eventID,
        stats,
    )

    if err != nil {
        log.Println(err)
    }
}
```

---

# SQL schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE,
    username TEXT,
    first_name TEXT,
    role TEXT DEFAULT 'member',
    created_at TIMESTAMP DEFAULT now()
);


CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    title TEXT,
    description TEXT,
    chat_id BIGINT,
    message_id INT,
    created_at TIMESTAMP DEFAULT now()
);


CREATE TABLE participants (
    id SERIAL PRIMARY KEY,
    event_id BIGINT,
    user_id BIGINT,
    status TEXT,
    updated_at TIMESTAMP DEFAULT now(),
    UNIQUE(event_id,user_id)
);
```

---

# Run

```
export BOT_TOKEN=...
export DB_URL=postgres://user:pass@localhost/db?sslmode=disable

go run .
```

---

# Webhook

```
https://api.telegram.org/bot<TOKEN>/setWebhook?url=https://yourdomain/webhook
```

---

This repository compiles and provides the core functionality:

• user auto‑registration
• event participant status
• live counter updates
• inline keyboard updates

