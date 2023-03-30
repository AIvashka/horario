package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Event struct {
	gorm.Model
	Name        string
	Description string
	Date        time.Time
	Note        string
	ChatID      int64
	Location    *time.Location
}

func (db *DB) AddEvent(event *Event) error {
	if err := db.Create(&event).Error; err != nil {
		return fmt.Errorf("failed to create event %v", event)
	}
	return nil
}

func (db *DB) GetEventsForToday(chatID int64, location *time.Location) []Event {
	now := time.Now().In(location).Unix()
	end := time.Unix(now, 0).AddDate(0, 0, 1).Unix()
	return db.GetEvents(now, end, chatID)
}

func (db *DB) GetEventsForTomorrow(chatID int64, location *time.Location) []Event {
	tomorrow := time.Now().In(location).AddDate(0, 0, 1).Unix()
	tomorrowEnd := time.Unix(tomorrow, 0).AddDate(0, 0, 1).Unix()
	return db.GetEvents(tomorrow, tomorrowEnd, chatID)
}

func (db *DB) GetEvents(start, end int64, chatID int64) []Event {
	startTime := time.Unix(start, 0).Truncate(24 * time.Hour)
	endTime := time.Unix(end, 0).Truncate(24 * time.Hour).Add(24 * time.Hour)
	var events []Event
	db.Where("date BETWEEN ? AND ? AND chat_id = ?", startTime, endTime, chatID).Find(&events)

	return events
}

func (db *DB) GetEventsForWeek(chatID int64, location *time.Location) []Event {
	now := time.Now().In(location).Unix()
	fmt.Println(now)
	weekEnd := time.Unix(now, 0).AddDate(0, 0, 7).Unix()
	return db.GetEvents(now, weekEnd, chatID)
}

func (db *DB) GetEventByID(id int) (*Event, error) {
	var event Event
	err := db.First(&event, id).Error
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (db *DB) DeleteEventByID(id int) error {
	event := Event{}
	err := db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return err
	}

	return db.Delete(&event).Error
}

func formatTasks(events []Event, includeDate bool, location *time.Location) string {
	var sb strings.Builder
	var format string
	if includeDate {
		format = "02-01-2006 15:04"
	} else {
		format = "15:04"
	}

	for _, event := range events {
		localDate := event.Date.In(location).Format(format)
		sb.WriteString(fmt.Sprintf("• %s %s [%d]\n", event.Name, localDate, event.Model.ID))
		if event.Note != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", event.Note))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func FormatWithTitle(title string, events []Event, includeDate bool, location *time.Location) string {
	return title + formatTasks(events, includeDate, location)
}
