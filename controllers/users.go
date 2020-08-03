package controllers

import (
	"github.com/backend-trainee-assignment/models"
	"github.com/backend-trainee-assignment/views"
	"net/http"
)

type Users struct {
	us models.UserService
}

func NewUsers(us models.UserService) *Users {
	return &Users{
		us: us,
	}
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := decodeJSONBody(w, r, &user)
	if err != nil {
		classificateErrorAndRenderView(w, err)
		return
	}

	result, statusCode, err := u.us.Create(&user)
	if err != nil {
		views.RenderJSON(w, nil, statusCode, err)
		return
	}

	views.RenderJSON(w, result, statusCode, nil)

	return
}
