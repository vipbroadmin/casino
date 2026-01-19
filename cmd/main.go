package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"players_service/internal/config"
	playerhttp "players_service/internal/delivery/http/player"
	"players_service/internal/infra/clock"
	"players_service/internal/infra/migrations"
	"players_service/internal/infra/postgres"
	outboxpg "players_service/internal/repository/outbox/postgres"
	playerpg "players_service/internal/repository/player/postgres"
	playeruc "players_service/internal/usecase/player"
)

func main() {
	// ===== config =====
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// ===== db =====
	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	// ===== migrations =====
	migrations.RunWithLog(db, "./migrations")

	// ===== infra =====
	uow := postgres.NewUnitOfWork(db)

	playerRepo := playerpg.New(db)
	eventRepo := playerpg.NewEvents(db)
	outboxRepo := outboxpg.New(db)
	documentsRepo := playerpg.NewDocuments(db)
	requisitesRepo := playerpg.NewRequisites(db)

	// ===== usecase =====
	playerService := playeruc.New(
		uow,
		playerRepo,
		eventRepo,
		outboxRepo,
		clock.New(),
		documentsRepo,
		requisitesRepo,
	)

	// ===== http =====
	handler := playerhttp.New(playerService)
	router := playerhttp.Routes(handler)

	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ===== graceful shutdown =====
	go func() {
		log.Printf("players-service started on :%s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("bye")
}
