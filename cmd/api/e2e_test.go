package main

import (
	"encoding/json"
	"goAPI/internal/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestWalletE2EFlow(t *testing.T) {
	app, cleanup := setupTestDB(t)
	defer cleanup()

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	// Создание кошелька и извлечение UUID из Location header
	var walletUUID string

	t.Run("Create Wallet", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/wallets", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		assert.Equal(t, resp.StatusCode, http.StatusCreated)

		location := resp.Header.Get("Location")
		if location == "" {
			t.Fatal("Location header is missing")
		}
		parts := strings.Split(location, "/")
		if len(parts) != 5 {
			t.Fatalf("unexpected Location format: %s", location)
		}
		walletUUID = parts[4]

		if _, err := uuid.Parse(walletUUID); err != nil {
			t.Fatalf("invalid UUID in Location: %s", walletUUID)
		}
	})

	// Проверка баланса (должен быть 0)
	t.Run("Show Balance", func(t *testing.T) {
		code, body := doRequest(t, ts, http.MethodGet, "/api/v1/wallets/"+walletUUID, nil)
		assert.Equal(t, code, http.StatusOK)

		var response struct {
			Wallet struct {
				Balance float64 `json:"balance"`
			} `json:"wallet"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("failed to decode response: %v. Raw body: %s", err, string(body))
		}
		assert.Equal(t, response.Wallet.Balance, 0.0)
	})

	//Депозит 100.50
	t.Run("Deposit", func(t *testing.T) {
		payload := map[string]any{
			"walletId":      walletUUID,
			"operationType": "DEPOSIT",
			"amount":        100.50,
		}
		code, body := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusOK)

		var response struct {
			Wallet struct {
				Balance float64 `json:"balance"`
			} `json:"wallet"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		assert.Equal(t, response.Wallet.Balance, 100.50)
	})

	// Снятие 50.00 -> баланс 50.50
	t.Run("Withdraw", func(t *testing.T) {
		payload := map[string]any{
			"walletId":      walletUUID,
			"operationType": "WITHDRAW",
			"amount":        50.00,
		}
		code, body := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusOK)

		var response struct {
			Wallet struct {
				Balance float64 `json:"balance"`
			} `json:"wallet"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		assert.Equal(t, response.Wallet.Balance, 50.50)
	})

	// Попытка снять больше, чем есть (должна быть ошибка валидации)
	t.Run("Withdraw Overdraft", func(t *testing.T) {
		payload := map[string]any{
			"walletId":      walletUUID,
			"operationType": "WITHDRAW",
			"amount":        100.00,
		}
		code, _ := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	})

	// Удаление кошелька
	t.Run("Delete Wallet", func(t *testing.T) {
		code, _ := doRequest(t, ts, http.MethodDelete, "/api/v1/wallets/"+walletUUID, nil)
		assert.Equal(t, code, http.StatusOK)
	})

	// Проверка, что удалённый кошелёк недоступен
	t.Run("Get Deleted Wallet", func(t *testing.T) {
		code, _ := doRequest(t, ts, http.MethodGet, "/api/v1/wallets/"+walletUUID, nil)
		assert.Equal(t, code, http.StatusNotFound)
	})

	// Негативные сценарии
	t.Run("Invalid UUID Format", func(t *testing.T) {
		code, _ := doRequest(t, ts, http.MethodGet, "/api/v1/wallets/invalid-uuid", nil)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	})

	t.Run("Change Balance with Invalid Operation", func(t *testing.T) {
		payload := map[string]any{
			"walletId":      walletUUID,
			"operationType": "UNKNOWN",
			"amount":        10,
		}
		code, _ := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	})

	t.Run("Change Balance with Zero Amount", func(t *testing.T) {
		payload := map[string]any{
			"walletId":      walletUUID,
			"operationType": "DEPOSIT",
			"amount":        0,
		}
		code, _ := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	})

	t.Run("Change Balance for Non-existent Wallet", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		payload := map[string]any{
			"walletId":      nonExistentUUID,
			"operationType": "DEPOSIT",
			"amount":        10,
		}
		code, _ := doRequest(t, ts, http.MethodPost, "/api/v1/wallet", payload)
		assert.Equal(t, code, http.StatusNotFound)
	})
}
