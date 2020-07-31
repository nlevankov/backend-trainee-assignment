package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
)

type ServicesConfig func(*Services) error

type Services struct {
	User    UserService
	Chat    ChatService
	Message MessageService

	db *gorm.DB
}

func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {

		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			log.Println("Can't connect to db, check your connection info in .config (if provided) or default connection info.")
			return err
		}
		s.db = db

		return nil
	}
}

func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)

		return nil
	}
}

func WithUser() ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db)
		return nil
	}
}

func WithChat() ServicesConfig {
	return func(s *Services) error {
		s.Chat = NewChatService(s.db)
		return nil
	}
}

func WithMessage() ServicesConfig {
	return func(s *Services) error {
		s.Message = NewMessageService(s.db)
		return nil
	}
}

func WithSetSchema(mode bool) ServicesConfig {
	return func(s *Services) error {

		if mode {
			setSchema(s.db)
		}

		return nil
	}
}

func (s *Services) Close() error {
	return s.db.Close()
}

func setSchema(db *gorm.DB) {

	db.Debug().Exec("DROP SCHEMA public CASCADE")
	db.Debug().Exec("CREATE SCHEMA public")

	db.CreateTable(
		&User{},
		&Chat{},
		&Message{},
	)

	db.Model(&Message{}).AddForeignKey("chat_id", "chats(id)", "CASCADE", "CASCADE")
	db.Model(&Message{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}
