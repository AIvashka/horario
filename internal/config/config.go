package config

import (
	"github.com/spf13/viper"
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
	// Set the default values
	viper.SetDefault("DBHost", "localhost")
	viper.SetDefault("DBPort", "5432")
	viper.SetDefault("DBUser", "postgres")
	viper.SetDefault("DBPassword", "")
	viper.SetDefault("DBName", "horario")
	viper.SetDefault("BotToken", "")

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

	// Read the environment variables
	viper.AutomaticEnv()

	// Override the config values with environment variables
	viper.SetEnvPrefix("HORARIO")
	viper.BindEnv("DBHost")
	viper.BindEnv("DBPort")
	viper.BindEnv("DBUser")
	viper.BindEnv("DBPassword")
	viper.BindEnv("DBName")
	viper.BindEnv("BotToken")

	// Create the config struct
	cfg := &Config{
		DBHost:     viper.GetString("DBHost"),
		DBPort:     viper.GetString("DBPort"),
		DBUser:     viper.GetString("DBUser"),
		DBPassword: viper.GetString("DBPassword"),
		DBName:     viper.GetString("DBName"),
		BotToken:   viper.GetString("BotToken"),
	}

	return cfg, nil
}
