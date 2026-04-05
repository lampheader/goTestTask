package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goAPI/internal/assert"
	"goAPI/internal/data"

	"github.com/google/uuid"
)

func TestWriteJSON(t *testing.T) {
	app := &application{}

	t.Run("Valid JSON with data.Wallet", func(t *testing.T) {
		rr := httptest.NewRecorder()

		walletID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		wallet := data.Wallet{
			UUID:    walletID,
			Balance: 250.75,
			Version: 1,
		}

		responseData := envelope{
			"wallet": wallet,
		}

		headers := make(http.Header)
		headers.Set("X-Custom-Header", "TestValue")

		err := app.writeJSON(rr, http.StatusOK, responseData, headers)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assert.Equal(t, rr.Code, http.StatusOK)
		assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")

		expectedBody := `{"wallet":{"uuid":"123e4567-e89b-12d3-a456-426614174000","balance":250.75,"version":1}}` + "\n"
		assert.Equal(t, strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expectedBody))
	})

	t.Run("JSON Marshal Error", func(t *testing.T) {
		rr := httptest.NewRecorder()

		unmarshalableData := envelope{
			"channel": make(chan int),
		}

		err := app.writeJSON(rr, http.StatusOK, unmarshalableData, nil)

		if err == nil {
			t.Error("expected an error when marshaling unmarshalable data, got nil")
		}
	})
}

func TestReadJSON(t *testing.T) {
	app := &application{}

	testUUID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name        string
		body        string
		wantError   bool
		errContains string
	}{
		{
			name:      "Valid JSON to data.Wallet",
			body:      `{"uuid": "` + testUUID + `", "balance": 1000.50, "version": 2}`,
			wantError: false,
		},
		{
			name:        "Syntax Error (missing quote)",
			body:        `{"uuid": "` + testUUID + `, "balance": 1000.50}`,
			wantError:   true,
			errContains: "body contains badly-formed JSON",
		},
		{
			name:        "Type Error (string instead of float64 for balance)",
			body:        `{"uuid": "` + testUUID + `", "balance": "one thousand"}`,
			wantError:   true,
			errContains: "body contains incorrect JSON type for field",
		},
		{
			name:        "Type Error (invalid UUID string)",
			body:        `{"uuid": "not-a-valid-uuid", "balance": 1000.50}`,
			wantError:   true,
			errContains: "invalid UUID",
		},
		{
			name:        "Empty Body",
			body:        ``,
			wantError:   true,
			errContains: "body must not be empty",
		},
		{
			name:        "Unknown Field",
			body:        `{"uuid": "` + testUUID + `", "balance": 1000.50, "admin": true}`,
			wantError:   true,
			errContains: "unknown field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			var dst data.Wallet
			err := app.readJSON(rr, req, &dst)

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, but got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, but got: %v", err)
				}
				assert.Equal(t, dst.UUID.String(), testUUID)
				assert.Equal(t, dst.Balance, 1000.50)
				assert.Equal(t, dst.Version, 2)
			}
		})
	}
}

func TestReadJSON_BodyTooLarge(t *testing.T) {
	app := &application{}

	largeBody := `{"balance": ` + strings.Repeat("9", 1_048_576) + `}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))
	rr := httptest.NewRecorder()

	var dst data.Wallet
	err := app.readJSON(rr, req, &dst)

	if err == nil {
		t.Fatal("expected an error for body too large, got nil")
	}

	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) && !strings.Contains(err.Error(), "http: request body too large") {
		t.Errorf("expected MaxBytesError, got: %v", err)
	}
}
