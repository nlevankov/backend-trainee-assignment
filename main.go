package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nlevankov/backend-trainee-assignment/controllers"
	"github.com/nlevankov/backend-trainee-assignment/models"
	"github.com/nlevankov/backend-trainee-assignment/views"
)

func main() {

	// flags' initialization

	setSchemaFlagPtr := flag.Bool("setschema", false, "WARNING: it is destructive action. Provide this flag "+
		"to initialize the storage.")

	flag.Parse()

	// the app's config's initialization

	cfg := LoadConfig()

	// creating services

	services, err := models.NewServices(
		models.WithGorm(cfg.Database.Dialect(), cfg.Database.ConnectionInfo(), int(cfg.StorageConnNumOfAttempts), cfg.StorageConnIntervalBWAttempts),
		models.WithLogMode(cfg.Logmode),
		models.WithUser(),
		models.WithChat(),
		models.WithMessage(),
		models.WithSetSchema(*setSchemaFlagPtr),
	)
	must(err)
	defer services.Close()

	r := mux.NewRouter()

	// initializing controllers

	usersC := controllers.NewUsers(services.User)
	chatsC := controllers.NewChats(services.Chat)
	messageC := controllers.NewMessages(services.Message)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		views.RenderJSON(w, nil, http.StatusNotFound, models.ErrNoSuchEndpointExists)
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		views.RenderJSON(w, nil, http.StatusNotFound, models.ErrNoSuchHTTPMethod)
	})
	r.HandleFunc("/users/add", usersC.Create).Methods(http.MethodPost)
	r.HandleFunc("/chats/add", chatsC.Create).Methods(http.MethodPost)
	r.HandleFunc("/messages/add", messageC.Create).Methods(http.MethodPost)
	r.HandleFunc("/chats/get", chatsC.ByUserID).Methods(http.MethodPost)
	r.HandleFunc("/messages/get", messageC.ByChatID).Methods(http.MethodPost)

	addr := fmt.Sprintf(cfg.IP+":%d", cfg.Port)
	go func() {
		must(http.ListenAndServe(addr, r))
	}()

	var n sync.WaitGroup
	n.Add(1)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigs
		log.Printf("Got <%v> signal, shutting down...", sig)
		n.Done()
	}()

	fmt.Printf("Started HTTP server on %v\nSend SIGINT or SIGTERM to exit\n", addr)

	n.Wait()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
