package controllers

import (
	"github.com/backend-trainee-assignment/models"
	"github.com/backend-trainee-assignment/views"
	"net/http"
)

type Chats struct {
	cs models.ChatService
}

func NewChats(cs models.ChatService) *Chats {
	return &Chats{
		cs: cs,
	}
}

func (c *Chats) Create(w http.ResponseWriter, r *http.Request) {
	var cqp models.ChatQueryParams

	err := decodeJSONBody(w, r, &cqp)
	if err != nil {
		classificateErrorAndRenderView(w, err)
		return
	}

	result, statusCode, err := c.cs.Create(&cqp)
	if err != nil {
		views.RenderJSON(w, nil, statusCode, err)
		return
	}

	views.RenderJSON(w, result, statusCode, nil)

	return
}

func (c *Chats) ByUserID(w http.ResponseWriter, r *http.Request) {
	var cqp models.ChatQueryParams

	err := decodeJSONBody(w, r, &cqp)
	if err != nil {
		classificateErrorAndRenderView(w, err)
		return
	}

	result, statusCode, err := c.cs.ByUserID(cqp.UserID)
	if err != nil {
		views.RenderJSON(w, nil, statusCode, err)
		return
	}

	views.RenderJSON(w, result, statusCode, nil)

	return
}
