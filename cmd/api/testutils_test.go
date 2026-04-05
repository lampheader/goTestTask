package main

import (
	"bytes"
	"context"
	"encoding/json"
	"goAPI/internal/data"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/julienschmidt/httprouter"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newTestApplication() *application {
	return &application{
		logger: log.New(io.Discard, "", 0),
		models: data.NewMockModels(),
	}
}

func testRoute(app *application, method, path string, handler http.HandlerFunc) *httprouter.Router {
	router := httprouter.New()
	router.HandlerFunc(method, path, handler)
	return router
}

func doRequest(t *testing.T, ts *httptest.Server, method, path string, body any) (int, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ts.URL+path, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, respBody
}

func setupTestDB(t *testing.T) (*application, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	m, err := migrate.New("file://../../migrations", connStr)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %s", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %s", err)
	}

	cfg := config{}
	cfg.db.dsn = connStr
	cfg.db.maxOpenConns = 25
	cfg.db.minIdleConns = 25
	cfg.db.maxIdleTime = "1m"

	db, err := openDB(cfg)
	if err != nil {
		t.Fatalf("failed to open db: %s", err)
	}

	app := &application{
		config: cfg,
		logger: log.New(io.Discard, "", 0),
		models: data.NewModels(db),
	}

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	return app, cleanup
}
