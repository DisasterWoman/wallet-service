package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DisasterWoman/wallet-service/internal/config"
	"github.com/DisasterWoman/wallet-service/internal/handler"
	"github.com/DisasterWoman/wallet-service/internal/repository"
	"github.com/DisasterWoman/wallet-service/internal/service"
	_ "github.com/DisasterWoman/wallet-service/docs" 
	"github.com/gorilla/mux"
	_"github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Wallet Service API
// @version 1.0
// @description REST API для управления виртуальными кошельками с поддержкой конкурентных операций

// @contact.name API Support
// @contact.url https://github.com/DisasterWoman/wallet-service

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080

// healthHandler обрабатывает запросы проверки здоровья
// @Summary Проверка здоровья сервиса
// @Description Возвращает статус работы сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Сервис работает"
// @Router /health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ok", "message": "Wallet service is running"}`)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to database: %s", cfg.DBName)

	repo := repository.NewPostgresRepository(db)
	walletService := service.NewWalletService(repo)
	walletHandler := handler.NewWalletHandler(walletService)

	r := mux.NewRouter()
	
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)                           
	r.HandleFunc("/api/v1/wallet", walletHandler.UpdateWalletBalance).Methods(http.MethodPost)  
	r.HandleFunc("/api/v1/wallets/{walletId}", walletHandler.GetWalletBalance).Methods(http.MethodGet)
	
	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: r,
	}

	go func() {
		log.Printf("Server started on %s", cfg.GetServerAddress())
		log.Printf("Swagger documentation available at: http://%s/swagger/index.html", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}