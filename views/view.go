package views

import (
	"encoding/json"
	"log"
	"net/http"
)

type PublicError interface {
	error
	Public() string
}

func RenderJSON(w http.ResponseWriter, result interface{}, StatusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)

	var msg *string
	if err != nil {
		if pErr, ok := err.(PublicError); ok {
			m := pErr.Public()
			msg = &m
		} else {
			log.Println(err)
		}
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	d := map[string]interface{}{"Result": result, "Error": msg}
	if err = enc.Encode(d); err != nil {
		log.Println(err)
	}
}
