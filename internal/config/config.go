package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"os"
)

// Config is the configuration struct
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	BotToken   string
}

// LoadConfig loads the configuration from environment variables and/or config file
func LoadConfig() (*Config, error) {

	// Look for .env file
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	v := os.Getenv("DBHOST")
	c := os.Getenv("DBPORT")
	fmt.Println(v, c)

	// Set the config file name and path
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	// Bind environment variables to viper
	viper.BindEnv("DBHOST")
	viper.BindEnv("DBPORT")
	viper.BindEnv("DBUSER")
	viper.BindEnv("DBPASSWORD")
	viper.BindEnv("DBNAME")
	viper.BindEnv("BOTTOKEN")

	// Read the values from viper
	cfg := &Config{
		DBHost:     viper.GetString("DBHOST"),
		DBPort:     viper.GetString("DBPORT"),
		DBUser:     viper.GetString("DBUSER"),
		DBPassword: viper.GetString("DBPASSWORD"),
		DBName:     viper.GetString("DBNAME"),
		BotToken:   viper.GetString("BOTTOKEN"),
	}

	return cfg, nil
}
