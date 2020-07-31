package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

type Message struct {
	ID        *uint   `gorm:"primary_key"`
	ChatID    *uint   `json:"chat"`   // chat
	UserID    *uint   `json:"author"` // author
	Text      *string `gorm:"not null" json:"text"`
	CreatedAt *time.Time
}

// todo наверняка можно отрефакторить и сделать сообщения в более унифицированном формате
const (
	ErrMessageChatDoesntExist modelError = "models: the chat with the provided id doesn't exist"
	ErrMessageUserDoesntExist modelError = "models: the user with the provided id doesn't exist"
	ErrMessageUserIsNotInChat modelError = "models: the user is not in the chat"

	ErrMessageChatIsNull   modelError = "models: 'chat' can't be null"
	ErrMessageAuthorIsNull modelError = "models: 'author' can't be null"
	ErrMessageTextIsNull   modelError = "models: 'text' can't be null"
	ErrMessageTextIsEmpty  modelError = "models: 'text' can't be empty"
)

type MessageService interface {
	MessageDB
}

type MessageDB interface {
	Create(msg *Message) (uint, int, error)
	ByChatID(chatid *uint) ([]*Message, int, error)
}

var _ MessageService = &messageService{}

type messageService struct {
	MessageDB
}

func NewMessageService(db *gorm.DB) MessageService {
	mg := &messageGorm{
		db: db,
	}

	mv := newMessageValidator(mg)

	return &messageService{
		MessageDB: mv,
	}
}

var _ MessageDB = &messageGorm{}

type messageGorm struct {
	db *gorm.DB
}

func (mg *messageGorm) Create(msg *Message) (uint, int, error) {
	var chat Chat
	err := mg.db.Where("id = ?", msg.ChatID).First(&chat).Error
	if chat.ID == nil {
		return 0, http.StatusNotFound, ErrMessageChatDoesntExist
	}

	var user User
	err = mg.db.Where("id = ?", msg.UserID).First(&user).Error
	if user.ID == nil {
		return 0, http.StatusNotFound, ErrMessageUserDoesntExist
	}

	err = mg.db.Preload("Users", "id = ?", user.ID).First(&chat).Error
	if len(chat.Users) == 0 {
		return 0, http.StatusUnauthorized, ErrMessageUserIsNotInChat
	}

	err = mg.db.Create(msg).Error
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	return *msg.ID, http.StatusOK, nil
}

func (mg *messageGorm) ByChatID(chatid *uint) ([]*Message, int, error) {
	var chat User
	err := mg.db.Where("id = ?", *chatid).First(&chat).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, ErrMessageChatDoesntExist
		}
		return nil, http.StatusInternalServerError, err
	}

	var msgs []*Message
	err = mg.db.
		Where("chat_id = ?", *chatid).
		Order("created_at").
		Find(&msgs).
		Error

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if len(msgs) == 0 {
		return nil, http.StatusOK, nil
	}

	return msgs, http.StatusOK, nil
}

type messageValidator struct {
	MessageDB
}

func newMessageValidator(mdb MessageDB) *messageValidator {
	return &messageValidator{
		MessageDB: mdb,
	}
}

func (mv *messageValidator) Create(msg *Message) (uint, int, error) {
	statusCode, err := runMessageValFns(msg,
		mv.messageChatNotNull,
		mv.messageAuthorNotNull,
		mv.messageTextNotNull,
		mv.messageTextNotEmpty,
	)
	if err != nil {
		return 0, statusCode, err
	}

	return mv.MessageDB.Create(msg)
}

func (mv *messageValidator) ByChatID(chatid *uint) ([]*Message, int, error) {
	statusCode, err := runMessageValFns(&Message{ChatID: chatid},
		mv.messageChatNotNull,
	)
	if err != nil {
		return nil, statusCode, err
	}

	return mv.MessageDB.ByChatID(chatid)
}

// валидаторы и нормализаторы

type messageValFn func(msg *Message) (int, error)

func runMessageValFns(msg *Message, fns ...messageValFn) (int, error) {
	for _, fn := range fns {
		statusCode, err := fn(msg)
		if err != nil {
			return statusCode, err
		}
	}
	return http.StatusOK, nil
}

func (mv *messageValidator) messageChatNotNull(msg *Message) (int, error) {
	if msg.ChatID == nil {
		return http.StatusBadRequest, ErrMessageChatIsNull
	}
	return http.StatusOK, nil
}

func (mv *messageValidator) messageAuthorNotNull(msg *Message) (int, error) {
	if msg.UserID == nil {
		return http.StatusBadRequest, ErrMessageAuthorIsNull
	}
	return http.StatusOK, nil
}

func (mv *messageValidator) messageTextNotNull(msg *Message) (int, error) {
	if msg.Text == nil {
		return http.StatusBadRequest, ErrMessageTextIsNull
	}
	return http.StatusOK, nil
}

func (mv *messageValidator) messageTextNotEmpty(msg *Message) (int, error) {
	if *msg.Text == "" {
		return http.StatusBadRequest, ErrMessageTextIsEmpty
	}
	return http.StatusOK, nil
}
