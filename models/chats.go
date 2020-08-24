package models

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"net/http"
	"strconv"
	"time"
)

type Chat struct {
	ID        *uint   `gorm:"primary_key"`
	Name      *string `gorm:"unique;not null"`
	Users     []*User `gorm:"many2many:chats_users"`
	CreatedAt *time.Time
	Messages  []*Message
}

type ChatQueryParams struct {
	Name    *string     `json:"name"`
	UserIDs []*stringID `json:"users"`
	UserID  *uint       `json:"user,string"`
}

// A helper type, just because ",string" struct tag doesn't work with slices
type stringID uint

func (sID *stringID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return err
	}

	*sID = stringID(v)

	return nil
}

const (
	ErrChatNameIsEmpty modelError = "'name' can't be empty"
	ErrChatNameIsNull  modelError = "'name' can't be null"
	ErrChatUserIsNull  modelError = "'user' can't be null"

	ErrChatUsersIsNull     modelError = "'users' can't be null"
	ErrChatUsersIsEmpty    modelError = "'users' can't be empty"
	ErrChatUsersIDsAreNull modelError = "'users' can't contain null(s)"

	ErrChatAlreadyExists      modelError = "The chat with this name already exists"
	ErrChatSomeUsersDontExist modelError = "Some users don't exist"
)

type ChatService interface {
	ChatDB
}

type ChatDB interface {
	Create(cqp *ChatQueryParams) (uint, int, error)
	ByUserID(userID *uint) ([]*Chat, int, error)
}

var _ ChatService = &chatService{}

type chatService struct {
	ChatDB
}

func NewChatService(db *gorm.DB) ChatService {
	cg := &chatGorm{
		db: db,
	}

	cv := newChatValidator(cg)

	return &chatService{
		ChatDB: cv,
	}
}

var _ ChatDB = &chatGorm{}

type chatGorm struct {
	db *gorm.DB
}

func (cg *chatGorm) Create(cqp *ChatQueryParams) (uint, int, error) {
	chat := &Chat{}
	err := cg.db.Where("name = ?", cqp.Name).First(&chat).Error
	if chat.ID != nil {
		return 0, http.StatusConflict, ErrChatAlreadyExists
	}

	var users []*User
	err = cg.db.Where("id in (?)", cqp.UserIDs).Find(&users).Error
	if len(users) != len(cqp.UserIDs) {
		return 0, http.StatusConflict, ErrChatSomeUsersDontExist
	}

	chat.Name = cqp.Name
	chat.Users = users
	err = cg.db.Create(&chat).Error
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	return *chat.ID, http.StatusOK, nil
}

func (cg *chatGorm) ByUserID(userID *uint) ([]*Chat, int, error) {
	var user User
	err := cg.db.Where("id = ?", *userID).First(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, ErrMessageUserDoesntExist
		}
		return nil, http.StatusInternalServerError, err
	}

	var chatIDs []uint
	var chats []*Chat

	// получаем id чатов пользователя в требуемом порядке
	// левое соединение нужно на случай отсутствия сообщений в чате.
	// orm генерирует несколько запросов для своих нужд, возможно неоптимально
	query := `SELECT id FROM (SELECT max(messages.created_at), chats.id
			FROM chats
			LEFT JOIN messages on chats.id = messages.chat_id
			WHERE chats.id in
			(SELECT chats_users.chat_id FROM chats_users WHERE chats_users.user_id = ?)
			GROUP BY chats.id
			ORDER BY max DESC NULLS LAST) as tempp`
	err = cg.db.
		Raw(query, *userID).
		Pluck("id", &chatIDs).Error
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// чтобы не делать лишних запросов далее, в случае если юзер не состоит ни в каких чатах
	if chatIDs == nil {
		return nil, http.StatusOK, nil
	}

	// получаем чаты пользователя с отсортированными сообщениями в них от позднего к раннему
	// todo можно ли сделать так, чтоб выражение "where in" выбирало в том порядке, что имеется в chatIDs?
	//  если да, то можно тогда сделать одним запросом то, что ниже
	for i := range chatIDs {
		var chat Chat
		err = cg.db.
			Preload("Users").
			Preload("Messages", func(db *gorm.DB) *gorm.DB {
				return db.Order("messages.created_at DESC")
			}).
			Where("id = ?", chatIDs[i]).First(&chat).
			Error

		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		chats = append(chats, &chat)
	}

	return chats, http.StatusOK, nil
}

type chatValidator struct {
	ChatDB
}

func newChatValidator(cdb ChatDB) *chatValidator {
	return &chatValidator{
		ChatDB: cdb,
	}
}

func (cv *chatValidator) Create(cqp *ChatQueryParams) (uint, int, error) {
	statusCode, err := runChatValFns(cqp,
		cv.chatNameNotNull,
		cv.chatUsersNotNull,
		cv.chatNameNotEmpty,
		cv.chatUsersNotEmpty,
		cv.chatUsersIDsNotNull,
		cv.chatUsersRemoveDuplicates)

	if err != nil {
		return 0, statusCode, err
	}

	return cv.ChatDB.Create(cqp)
}

func (cv *chatValidator) ByUserID(userID *uint) ([]*Chat, int, error) {
	cqp := ChatQueryParams{UserID: userID}
	statusCode, err := runChatValFns(&cqp,
		cv.chatUserNotNull)
	if err != nil {
		return nil, statusCode, err
	}

	return cv.ChatDB.ByUserID(userID)
}

type chatValFn func(params *ChatQueryParams) (int, error)

func runChatValFns(cqv *ChatQueryParams, fns ...chatValFn) (int, error) {
	for _, fn := range fns {
		statusCode, err := fn(cqv)
		if err != nil {
			return statusCode, err
		}
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatUserNotNull(cqv *ChatQueryParams) (int, error) {
	if cqv.UserID == nil {
		return http.StatusBadRequest, ErrChatUserIsNull
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatNameNotEmpty(cqv *ChatQueryParams) (int, error) {
	if *cqv.Name == "" {
		return http.StatusBadRequest, ErrChatNameIsEmpty
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatNameNotNull(cqv *ChatQueryParams) (int, error) {
	if cqv.Name == nil {
		return http.StatusBadRequest, ErrChatNameIsNull
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatUsersNotNull(cqv *ChatQueryParams) (int, error) {
	if cqv.UserIDs == nil {
		return http.StatusBadRequest, ErrChatUsersIsNull
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatUsersNotEmpty(cqv *ChatQueryParams) (int, error) {
	if len(cqv.UserIDs) == 0 {
		return http.StatusBadRequest, ErrChatUsersIsEmpty
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatUsersIDsNotNull(cqv *ChatQueryParams) (int, error) {
	for i := range cqv.UserIDs {
		if cqv.UserIDs[i] == nil {
			return http.StatusBadRequest, ErrChatUsersIDsAreNull
		}
	}
	return http.StatusOK, nil
}

func (cv *chatValidator) chatUsersRemoveDuplicates(cqv *ChatQueryParams) (int, error) {
	seen := make(map[stringID]struct{})
	for _, item := range cqv.UserIDs {
		seen[*item] = struct{}{}
	}

	cqv.UserIDs = nil

	for item := range seen {
		func(id stringID) {
			cqv.UserIDs = append(cqv.UserIDs, &id)
		}(item)
	}

	return http.StatusOK, nil
}
