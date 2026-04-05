package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goAPI/internal/assert"
	"goAPI/internal/data"

	"github.com/google/uuid"
)

func TestCreateWalletHandler(t *testing.T) {
	app := newTestApplication()
	router := app.routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusCreated)

	if rr.Header().Get("Location") == "" {
		t.Error("expected Location header to be set")
	}

	bodyBytes := rr.Body.Bytes()

	var response map[string]string

	err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v. Raw body: %s", err, string(bodyBytes))
	}

	assert.Equal(t, response["message"], "wallet successfully created")
}

func TestShowWalletBalance(t *testing.T) {
	app := newTestApplication()
	router := app.routes()

	wallet, _ := app.models.Wallets.Create()

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"Valid UUID (Exists)", "/api/v1/wallets/" + wallet.UUID.String(), http.StatusOK},
		{"Valid UUID (Not Found)", "/api/v1/wallets/" + uuid.New().String(), http.StatusNotFound},
		{"Invalid UUID Format", "/api/v1/wallets/invalid-uuid", http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, rr.Code, tt.wantStatus)
		})
	}
}

func TestChangeWalletBalance(t *testing.T) {
	app := newTestApplication()
	router := app.routes()

	wallet, _ := app.models.Wallets.Create()

	tests := []struct {
		name        string
		payload     map[string]any
		wantStatus  int
		wantBalance float64
	}{
		{
			name: "Valid Deposit",
			payload: map[string]any{
				"walletId":      wallet.UUID.String(),
				"operationType": "DEPOSIT",
				"amount":        100.50,
			},
			wantStatus:  http.StatusOK,
			wantBalance: 100.50,
		},
		{
			name: "Valid Withdraw",
			payload: map[string]any{
				"walletId":      wallet.UUID.String(),
				"operationType": "WITHDRAW",
				"amount":        50.00,
			},
			wantStatus:  http.StatusOK,
			wantBalance: 50.50,
		},
		{
			name: "Invalid Operation Type",
			payload: map[string]any{
				"walletId":      wallet.UUID.String(),
				"operationType": "UNKNOWN",
				"amount":        10.0,
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Amount Zero",
			payload: map[string]any{
				"walletId":      wallet.UUID.String(),
				"operationType": "DEPOSIT",
				"amount":        0,
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Wallet Not Found",
			payload: map[string]any{
				"walletId":      uuid.New().String(),
				"operationType": "DEPOSIT",
				"amount":        10.0,
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, rr.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Wallet data.Wallet `json:"wallet"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				assert.Equal(t, response.Wallet.Balance, tt.wantBalance)
			}
		})
	}
}

func TestDeleteWallet(t *testing.T) {
	app := newTestApplication()
	router := app.routes()

	wallet, _ := app.models.Wallets.Create()

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"Valid Delete", "/api/v1/wallets/" + wallet.UUID.String(), http.StatusOK},
		{"Delete Not Found", "/api/v1/wallets/" + uuid.New().String(), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, rr.Code, tt.wantStatus)
		})
	}
}
