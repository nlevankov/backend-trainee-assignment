package controllers

import (
	"github.com/nlevankov/backend-trainee-assignment/models"
	"github.com/nlevankov/backend-trainee-assignment/views"
	"net/http"
)

type Message struct {
	ms models.MessageService
}

func NewMessages(ms models.MessageService) *Message {
	return &Message{
		ms: ms,
	}
}

// todo отрефакторить, во всех контроллерах
func (m *Message) Create(w http.ResponseWriter, r *http.Request) {
	var msg models.Message

	err := decodeJSONBody(w, r, &msg)
	if err != nil {
		classificateErrorAndRenderView(w, err)
		return
	}

	result, statusCode, err := m.ms.Create(&msg)
	if err != nil {
		views.RenderJSON(w, nil, statusCode, err)
		return
	}

	views.RenderJSON(w, result, statusCode, nil)

	return
}

func (m *Message) ByChatID(w http.ResponseWriter, r *http.Request) {
	var msg models.Message

	err := decodeJSONBody(w, r, &msg)
	if err != nil {
		classificateErrorAndRenderView(w, err)
		return
	}

	result, statusCode, err := m.ms.ByChatID(msg.ChatID)
	if err != nil {
		views.RenderJSON(w, nil, statusCode, err)
		return
	}

	views.RenderJSON(w, result, statusCode, nil)

	return
}
