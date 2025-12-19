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

	playerhttp "players_service/internal/delivery/http/player"
	"players_service/internal/infra/clock"
	"players_service/internal/infra/postgres"
	outboxpg "players_service/internal/repository/outbox/postgres"
	playerpg "players_service/internal/repository/player/postgres"
	playeruc "players_service/internal/usecase/player"
)

func main() {
	// ===== config =====
	httpPort := getenv("APP_HTTP_PORT", "8080")
	pgDSN := buildPostgresDSN()

	// ===== db =====
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	// ===== infra =====
	uow := postgres.NewUnitOfWork(db)

	playerRepo := playerpg.New(db)
	eventRepo := playerpg.NewEvents(db)
	outboxRepo := outboxpg.New(db)

	// ===== usecase =====
	playerService := playeruc.New(
		uow,
		playerRepo,
		eventRepo,
		outboxRepo,
		clock.New(),
	)

	// ===== http =====
	handler := playerhttp.New(playerService)
	router := playerhttp.Routes(handler)

	server := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ===== graceful shutdown =====
	go func() {
		log.Printf("players-service started on :%s", httpPort)
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

// --- helpers ---

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func buildPostgresDSN() string {
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	db := getenv("POSTGRES_DB", "players")
	user := getenv("POSTGRES_USER", "players")
	pass := getenv("POSTGRES_PASSWORD", "players")
	ssl := getenv("POSTGRES_SSLMODE", "disable")

	return "postgres://" + user + ":" + pass +
		"@" + host + ":" + port +
		"/" + db +
		"?sslmode=" + ssl
}
