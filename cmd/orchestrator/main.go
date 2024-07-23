package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tkerm94/scalable-calculator/internal/database"
	messagebroker "github.com/tkerm94/scalable-calculator/internal/message_broker"
	"github.com/tkerm94/scalable-calculator/internal/orchestrator"
)

func main() {
	if err := orchestrator.ParseTemplateDir("../../web/templates", "layout"); err != nil {
		log.Fatalf("Couldnt parse templates folder: %s", err)
	}
	port := 8080
	http_logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	http_logger.Println("Server is starting...")
	c := &orchestrator.Controller{
		Logger: http_logger,
		NextRequestID: func() string {
			return strconv.FormatInt(time.Now().UnixNano(), 36)
		},
	}
	router := http.NewServeMux()
	router.HandleFunc("/calculator", c.Calculator)
	router.HandleFunc("/settings", c.Settings)
	router.HandleFunc("/profile", c.Profile)
	router.HandleFunc("/login", c.Login)
	router.HandleFunc("/register", c.Register)
	router.HandleFunc("/api/v1/login", c.Login_api)
	router.HandleFunc("/api/v1/register", c.Register_api)
	router.HandleFunc("/favicon.ico", c.Favicon)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      (orchestrator.Middlewares{c.Tracing, c.Logging}).Apply(router),
		ErrorLog:     http_logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := c.Shutdown(context.Background(), server)
	http_logger.Printf("Server is running at %d\n", port)
	database.InitDB()
	defer database.DB.Close()
	go messagebroker.ConnectOrchestratorToRMQ()
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		http_logger.Fatalf("Error on %d: %s\n", port, err)
	}
	<-ctx.Done()
	log.Println("Disconnected from DB")
	http_logger.Println("Server stopped")
}
