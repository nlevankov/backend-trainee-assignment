package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Config struct {
	IP                            string `json:"ip"`
	Port                          uint   `json:"port"`
	StorageConnNumOfAttempts      uint   `json:"storage_conn_num_of_attempts"`
	StorageConnIntervalBWAttempts uint   `json:"storage_conn_interval_bw_attempts"`
	Logmode                       bool   `json:"log_mode"`

	Database PostgresConfig `json:"database"`
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

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "123",
		Name:     "bta_dev",
	}
}

func DefaultConfig() Config {
	return Config{
		Database:                      DefaultPostgresConfig(),
		IP:                            "",
		Port:                          9000,
		StorageConnNumOfAttempts:      3,
		StorageConnIntervalBWAttempts: 3,
		Logmode:                       false,
	}
}

func LoadConfig(configReq bool) Config {
	f, err := os.Open(".config")
	if err != nil {
		if configReq {
			fmt.Println("A .config file must be provided with the -prod flag, shutting down.")
			panic(err)
		}

		fmt.Println("Using the default config...")
		return DefaultConfig()
	}

	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}

	if c.Port == 0 {
		log.Fatal("HTTP port can't be 0")
	}
	if c.StorageConnNumOfAttempts == 0 {
		log.Fatal("A number of attempts can't be 0 (Storage reconnection parameter)")
	}
	if c.StorageConnIntervalBWAttempts == 0 {
		log.Fatal("An interval between attempts can't be 0 (Storage reconnection parameter)")
	}

	fmt.Println("Successfully loaded .config")
	return c
}
