package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"bibliotheca/internal/handler"
	"bibliotheca/internal/middleware"
	"bibliotheca/internal/repository"
	"bibliotheca/internal/seed"
	"bibliotheca/internal/service"
	"bibliotheca/migrations"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	var lastErr error
	for i := 0; i < 20; i++ { // the DB container may still be starting
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		lastErr = db.PingContext(ctx)
		cancel()
		if lastErr == nil {
			return db, nil
		}
		log.Printf("waiting for database (%d/20): %v", i+1, lastErr)
		time.Sleep(2 * time.Second)
	}
	return nil, lastErr
}

func migrate(db *sql.DB) error {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[len(e.Name())-4:] == ".sql" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, n := range names {
		b, err := migrations.FS.ReadFile(n)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("migration %s: %w", n, err)
		}
		log.Printf("applied migration %s", n)
	}
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Data is always stored in PostgreSQL (relational database).
	// The default matches docker-compose.yml; override with DATABASE_URL.
	url := env("DATABASE_URL", "postgres://bib:bib@localhost:5432/bibliotheca?sslmode=disable")
	db, err := connect(url)
	if err != nil {
		log.Fatalf("cannot connect to PostgreSQL: %v\n"+
			"Start the database first (docker compose up -d db) or set DATABASE_URL. See README.md", err)
	}
	defer db.Close()
	if err := migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	var store repository.Store = repository.NewPostgresStore(db)
	log.Println("using PostgreSQL")

	clock := service.Clock(time.Now)
	cfg := service.DefaultConfig()
	books := service.NewBookService(store.Books(), clock)
	members := service.NewMemberService(store.Members(), clock)
	loans := service.NewLoanService(store.Loans(), store.Books(), store.Members(), cfg, clock)
	dash := service.NewDashboardService(store, loans, clock)

	if env("SEED", "true") == "true" { // seeds only when the catalogue is empty
		seed.Run(ctx, store, books, members)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // background overdue watcher
		defer wg.Done()
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if n, err := store.Loans().CountOverdue(ctx, time.Now()); err == nil && n > 0 {
					log.Printf("overdue watcher: %d overdue loan(s)", n)
				}
			}
		}
	}()

	api := handler.New(books, members, loans, dash)
	h := middleware.Chain(api.Routes(), middleware.Recover, middleware.Logger, middleware.CORS(env("CORS_ORIGIN", "http://localhost:4200")))
	srv := &http.Server{Addr: ":" + env("PORT", "8080"), Handler: h,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second}

	go func() {
		log.Printf("Bibliotheca API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(sctx)
	wg.Wait()
	log.Println("bye")
}
