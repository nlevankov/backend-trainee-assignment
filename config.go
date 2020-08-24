package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
)

type PostgresConfig struct {
	Host     string `env:"APP_STORAGE_HOST" envDefault:"database"`
	Port     int    `env:"APP_STORAGE_PORT" envDefault:"5432"`
	User     string `env:"APP_STORAGE_USER" envDefault:"postgres"`
	Password string `env:"APP_STORAGE_PWD" envDefault:"123"`
	Name     string `env:"APP_STORAGE_DBNAME" envDefault:"bta_dev"`
}

type Config struct {
	IP                            string `env:"APP_IP" envDefault:""`
	Port                          uint   `env:"APP_PORT" envDefault:"9000"`
	StorageConnNumOfAttempts      uint   `env:"APP_RETRY_NUM" envDefault:"5"`
	StorageConnIntervalBWAttempts uint   `env:"APP_RETRY_INTERVAL" envDefault:"3"`
	Logmode                       bool   `env:"APP_LOGMODE" envDefault:"false"`

	Database PostgresConfig
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}
func (c PostgresConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s "+
			"sslmode=disable", c.Host, c.Port, c.User, c.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable", c.Host, c.Port, c.User,
		c.Password, c.Name)
}

func LoadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Can't parse env variables into config structure")
		panic(err)
	}

	if cfg.Port == 0 {
		log.Fatal("HTTP port can't be 0")
	}
	if cfg.StorageConnNumOfAttempts == 0 {
		log.Fatal("A number of attempts can't be 0 (Storage reconnection parameter)")
	}
	if cfg.StorageConnIntervalBWAttempts == 0 {
		log.Fatal("An interval between attempts can't be 0 (Storage reconnection parameter)")
	}

	fmt.Println("Successfully loaded .config")

	return cfg

}
