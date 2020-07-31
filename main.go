package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/backend-trainee-assignment/controllers"
	"github.com/backend-trainee-assignment/models"
)

func main() {

	// flags' initialization

	prodFlagPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts and enables log mode.")

	setSchemaFlagPtr := flag.Bool("setschema", false, "WARNING: it is destructive action. Provide this flag "+
		"to set the db schema. If '-prod' flag is provided, this flag will be ignored.")

	flag.Parse()

	// the app's config's initialization

	cfg := LoadConfig(*prodFlagPtr)
	dbCfg := cfg.Database

	// creating services

	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		// Only log when not in prod
		models.WithLogMode(*prodFlagPtr),
		models.WithUser(),
		models.WithChat(),
		models.WithMessage(),
		models.WithSetSchema(!(*prodFlagPtr) && *setSchemaFlagPtr),
	)
	must(err)
	defer services.Close()

	fmt.Println("Successfully connected!")

	r := mux.NewRouter()

	// initializing controllers

	usersC := controllers.NewUsers(services.User)
	chatsC := controllers.NewChats(services.Chat)
	messageC := controllers.NewMessages(services.Message)

	r.HandleFunc("/users/add", usersC.Create).Methods(http.MethodPost)
	r.HandleFunc("/chats/add", chatsC.Create).Methods(http.MethodPost)
	r.HandleFunc("/messages/add", messageC.Create).Methods(http.MethodPost)
	r.HandleFunc("/chats/get", chatsC.ByUserID).Methods(http.MethodPost)
	r.HandleFunc("/messages/get", messageC.ByChatID).Methods(http.MethodPost)

	must(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
