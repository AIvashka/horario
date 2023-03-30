package events

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"horario/internal/config"
)

type DB struct {
	*gorm.DB
}

func NewDB(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Event{})

	return &DB{db}, nil
}
