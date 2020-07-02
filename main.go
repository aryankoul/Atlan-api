package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aryankoul/atlan-assignment/handlers"
	"github.com/gorilla/mux"
)

func main() {
	logger := log.New(os.Stdout, "collect-api ", log.LstdFlags)
	var wg sync.WaitGroup
	router := mux.NewRouter()

	taskHandler := handlers.NewTaskHandler(logger, &wg)

	router.HandleFunc("/create", taskHandler.CreateTask).Methods("GET")

	api := router.PathPrefix("/").Subrouter()
	api.Use(taskHandler.MiddlewareCheckTask)

	api.HandleFunc("/pause/{id}", taskHandler.PauseTask).Methods("GET")
	api.HandleFunc("/delete/{id}", taskHandler.DeleteTask).Methods("GET")
	api.HandleFunc("/resume/{id}", taskHandler.ResumeTask).Methods("GET")

	s := &http.Server{
		Addr:    ":9090",
		Handler: router,
	}

	go func() {
		logger.Println("Starting server on port 9090")
		err := s.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}

	}()

	// listen for os interrupt and kill commands on server
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan

	// send kill signal to all tasks so that they can rollback gracefully before server is closed
	taskHandler.KillAllTask()

	// Used such that the go program waits for all the goroutines to finish before it closes
	wg.Wait()

	logger.Println("Recieved terminate, properly terminated all tasks")
	logger.Println("Reason:", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
