package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"github.com/nlevankov/backend-trainee-assignment/views"
)

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func (mr *malformedRequest) Public() string {
	return mr.Error()
}

// it assumes that err != nil
func classificateErrorAndRenderView(w http.ResponseWriter, err error) {
	var mr *malformedRequest
	if errors.As(err, &mr) {
		views.RenderJSON(w, nil, mr.status, err)
	} else {
		views.RenderJSON(w, nil, http.StatusInternalServerError, err)
	}
}

// возможно существует пакет, который реализует эти стандартные проверки
//  решение взял отсюда: https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var numError *strconv.NumError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the '%s' field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			fieldNameWOQuotes := strings.Trim(fieldName, "\"")
			msg := fmt.Sprintf("Request body contains unknown field '%s'", fieldNameWOQuotes)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

			//todo пока так, наверняка можно что-то по лучше придумать, но это по крайней мере лучше,
			//  чем 500-ая ошибка
		case errors.As(err, &numError):
			typeName := strings.TrimPrefix(numError.Func, "Parse")
			msg := fmt.Sprintf("Trying to parse '%s' into %v", numError.Num, typeName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

			//todo пока так, наверняка можно что-то по лучше придумать, но это по крайней мере лучше,
			//  чем 500-ая ошибка
		case strings.HasPrefix(err.Error(), "json: invalid use of ,string struct tag, "):
			msg := strings.TrimPrefix(err.Error(), "json: invalid use of ,string struct tag, ")
			msg = strings.ReplaceAll(msg, "\"", "'")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
