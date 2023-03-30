package bot

import (
	"fmt"
	ev "horario/internal/events"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	botAPI         *tgbotapi.BotAPI
	db             *ev.DB
	currentCommand string
	currentChatID  int64
	timers         map[int]*time.Timer // добавляем словарь для хранения таймеров

}

func NewBot(token string, db *ev.DB) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	if err != nil {
		return nil, err
	}

	return &Bot{
		botAPI: bot,
		db:     db,
		timers: make(map[int]*time.Timer),
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesChan := b.botAPI.GetUpdatesChan(u)

	for update := range updatesChan {
		if update.Message == nil {
			continue
		}

		b.handleUpdate(&update)

	}

	return nil
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	command := update.Message.Command()
	message := strings.TrimSpace(update.Message.Text)

	if command != "" {
		b.currentCommand = command
		b.currentChatID = chatID
	}

	switch b.currentCommand {
	case "start":
		b.handleStart(chatID)
	case "help":
		b.handleHelp(chatID)
	case "add":
		if message == "/add" {
			b.handleAdd(chatID)
		} else {
			b.handleAddMessage(update.Message)
		}
	case "today":
		b.handleToday(update.Message)
	case "tomorrow":
		b.handleTomorrow(update.Message)
	case "week":
		b.handleWeek(update.Message)
	case "delete":
		if message == "/delete" {
			b.handleDelete(chatID)
		} else {
			b.handleDeleteMessage(update.Message)
		}
	default:
		b.handleUnknown(chatID)
	}

}

func (b *Bot) handleDelete(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Пожалуйста введи ID события, которое нужно удалить:")
	b.botAPI.Send(msg)
}

func (b *Bot) handleDeleteMessage(msg *tgbotapi.Message) {
	message := removePrefix(msg.Text, "/delete")
	eventID, err := strconv.Atoi(message)
	if err != nil {
		b.handleDeleteError(msg.Chat.ID)
		return
	}

	// отменяем таймер, если он был установлен для этой задачи
	b.cancelReminder(eventID)

	err = b.db.DeleteEventByID(eventID)
	if err != nil {
		b.handleDeleteError(msg.Chat.ID)
		return
	}

	textMessage := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Событие %d успешно удалено!", eventID))
	b.botAPI.Send(textMessage)
}

func (b *Bot) handleDeleteError(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Извини, не могу удалить это событие, не верный ID")
	b.botAPI.Send(msg)
}

func (b *Bot) handleStart(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Привет, Я бот, который поможет тебе вести расписание и ни о чем не забывать! Введи команду /help, чтобы посмотреть, что я умею")
	b.botAPI.Send(msg)
}

func (b *Bot) handleHelp(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Вот, что я могу делать:\n\n"+
		"/add [Ежедневный звонок | 30-04-2023 15:00 | Рассказать об успехах за день] — добавляет задачу в календарь\n"+
		"/today - расписание на сегодня\n"+
		"/tomorrow - расписание на завтра\n"+
		"/week - расписание ближайшие 7 дней\n"+
		"/help - показывает это сообщение\n")
	b.botAPI.Send(msg)
}

func (b *Bot) handleAdd(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Добавь новое событие в формате [Звонок по работе|30-03-2023 15:00|Обсуждаем детали нового проекта]")
	b.botAPI.Send(msg)
}

func (b *Bot) handleAddMessage(msg *tgbotapi.Message) {
	message := removePrefix(msg.Text, "/add")

	parts := strings.Split(message, "|")
	var note string
	if len(parts) == 3 {
		note = strings.TrimSpace(parts[2])
	}
	name := strings.TrimSpace(parts[0])
	dateTimeStr := strings.TrimSpace(parts[1])

	date, err := time.ParseInLocation("02-01-2006 15:04", dateTimeStr, msg.Time().Location())
	if err != nil {
		b.handleAddError(msg.Chat.ID)
		return
	}

	event := ev.Event{Name: name, Date: date, Note: note, ChatID: msg.Chat.ID, Location: msg.Time().Location()}

	err = b.db.AddEvent(&event)
	if err != nil {
		log.Printf("%v", err)
		b.handleAddError(msg.Chat.ID)
		return
	}

	// schedule a reminder 30 minutes before the event
	reminderTime := event.Date.Add(-30 * time.Minute)
	go b.scheduleReminder(int(event.Model.ID), reminderTime, msg.Time().Location())

	textMessage := tgbotapi.NewMessage(msg.Chat.ID, "Событие добавлено в расписание!")
	b.botAPI.Send(textMessage)
}

func (b *Bot) handleToday(msg *tgbotapi.Message) {
	events := b.db.GetEventsForToday(msg.Chat.ID, msg.Time().Location())
	var text string
	if len(events) == 0 {
		text = "На сегодня мероприятий не запланировано! 🎉"
	} else {
		text = ev.FormatWithTitle("Расписание на сегодня 📅👇:\n\n", events, false, msg.Time().Location())
	}
	textMessage := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.botAPI.Send(textMessage)
}

func (b *Bot) handleTomorrow(msg *tgbotapi.Message) {
	events := b.db.GetEventsForTomorrow(msg.Chat.ID, msg.Time().Location())
	var text string
	if len(events) == 0 {
		text = "Нет мероприятий на завтра! 🎉"
	} else {
		text = ev.FormatWithTitle("Расписание на завтра 📅👇:\n\n", events, false, msg.Time().Location())
	}
	textMessage := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.botAPI.Send(textMessage)
}

func (b *Bot) handleWeek(msg *tgbotapi.Message) {
	events := b.db.GetEventsForWeek(msg.Chat.ID, msg.Time().Location())
	var text string
	if len(events) == 0 {
		text = "Нет мероприятий на ближайшие 7 дней! 🎉"
	} else {
		text = ev.FormatWithTitle("Расписание на неделю 📅👇:\n\n", events, true, msg.Time().Location())
	}
	textMessage := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.botAPI.Send(textMessage)
}

func (b *Bot) handleUnknown(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Извини, я не понял команду, попробуй использовать /help, чтобы узнать, что я могу")
	b.botAPI.Send(msg)
}

func (b *Bot) handleError(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, there was an error processing your request.")
	b.botAPI.Send(msg)
}

func (b *Bot) handleAddError(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, I couldn't add the event. Please enter the name and date of the event, separated by a space (dd-mm-yyyy). You can also add an optional note after the date.")
	b.botAPI.Send(msg)
}

func (b *Bot) scheduleReminder(eventID int, reminderTime time.Time, location *time.Location) {
	now := time.Now().In(location)
	if reminderTime.Before(now) {
		log.Printf("the reminder time has already passed")
		return
	}

	duration := reminderTime.Sub(now)

	timer := time.NewTimer(duration)
	// добавляем таймер в словарь
	b.timers[eventID] = timer

	log.Printf("reminder created for %v, with duration %v", reminderTime, duration.String())

	<-timer.C

	// удаляем таймер из словаря после срабатывания
	delete(b.timers, eventID)

	event, err := b.db.GetEventByID(eventID)
	fmt.Println(event)
	if err != nil {
		log.Printf("error occurred %v", err)
		return
	}

	msg := tgbotapi.NewMessage(event.ChatID, fmt.Sprintf("Напоминание! Событие: '%s' через 30 минут!", event.Name))
	b.botAPI.Send(msg)
}

func (b *Bot) cancelReminder(eventID int) {
	timer, ok := b.timers[eventID]
	if ok {
		if !timer.Stop() {
			<-timer.C
		}
		delete(b.timers, eventID)
	}
	log.Printf("timer was successfully cancelled")
}

func removePrefix(str, prefix string) string {
	if strings.HasPrefix(str, prefix) {
		part := strings.SplitN(str, " ", 2)
		return part[1]
	} else {
		return str
	}
}
