package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"net/http"
	"time"
)

type User struct {
	ID        *uint   `gorm:"primary_key"`
	Name      *string `gorm:"unique;not null" json:"username"` // username, префикс  user избыточен
	CreatedAt *time.Time
	Chats     []*Chat `gorm:"many2many:chats_users"`
	Messages  []*Message
	// возможно было бы оправдано иметь поле DeletedAt (для soft delete)
}

const (
	ErrUserNameIsEmpty modelError = "models: 'username' can't be empty"

	ErrUserNameIsNull modelError = "models: 'username' can't be null"

	ErrUserAlreadyExists modelError = "models: user with this name already exists"
)

type UserService interface {
	UserDB
}

type UserDB interface {
	Create(user *User) (uint, int, error)
}

var _ UserService = &userService{}

type userService struct {
	UserDB
}

func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{
		db: db,
	}

	uv := newUserValidator(ug)

	return &userService{
		UserDB: uv,
	}
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}

func (ug *userGorm) Create(user *User) (uint, int, error) {
	err := ug.db.Create(&user).Error

	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if e.Code == "23505" {
				return 0, http.StatusConflict, ErrUserAlreadyExists
			}
		}
		return 0, http.StatusInternalServerError, err

	}

	return *user.ID, http.StatusOK, nil
}

type userValidator struct {
	UserDB
}

func newUserValidator(udb UserDB) *userValidator {
	return &userValidator{
		UserDB: udb,
	}
}

func (uv *userValidator) Create(user *User) (uint, int, error) {
	statusCode, err := runUserValFns(user,
		uv.userNameNotNull,
		uv.userNameNotEmpty)
	if err != nil {
		return 0, statusCode, err
	}

	return uv.UserDB.Create(user)
}

type userValFn func(*User) (int, error)

func runUserValFns(user *User, fns ...userValFn) (int, error) {
	for _, fn := range fns {
		statusCode, err := fn(user)
		if err != nil {
			return statusCode, err
		}
	}
	return http.StatusOK, nil
}

func (uv *userValidator) userNameNotEmpty(user *User) (int, error) {
	if *user.Name == "" {
		return http.StatusBadRequest, ErrUserNameIsEmpty
	}
	return http.StatusOK, nil
}

func (uv *userValidator) userNameNotNull(user *User) (int, error) {
	if user.Name == nil {
		return http.StatusBadRequest, ErrUserNameIsNull
	}
	return http.StatusOK, nil
}
