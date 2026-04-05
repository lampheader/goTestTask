package main

import (
	"context"
	"flag"
	"fmt"
	"goAPI/internal/data"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const version = "1.0.0"

type config struct {
	port int
	db   struct {
		dsn          string
		maxOpenConns int
		minIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {

	var cfg config
	parseConfigInputFlags(&cfg)

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting server on %s", srv.Addr)

	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func parseConfigInputFlags(cfg *config) {

	Port := 4000 //default

	if envPort := os.Getenv("SRV_PORT"); envPort != "" {
		if parsedPort, err := strconv.Atoi(envPort); err == nil && parsedPort > 0 {
			Port = parsedPort
		}

	}

	flag.IntVar(&cfg.port, "port", Port, "API server port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("SERVICE_DATABASE_URL"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.minIdleConns, "db-min-idle-conns", 25, "PostgreSQL min idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "1m", "PostgreSQL max connection idle time")

	flag.Parse()
}

func openDB(cfg config) (*pgxpool.Pool, error) {

	db, err := pgxpool.New(context.Background(), cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.Config().MaxConns = int32(cfg.db.maxOpenConns)
	db.Config().MinIdleConns = int32(cfg.db.minIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.Config().MaxConnIdleTime = duration

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
