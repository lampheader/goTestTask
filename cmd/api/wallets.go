package main

import (
	"errors"
	"fmt"
	"goAPI/internal/data"
	"goAPI/internal/validator"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createWalletHandler(w http.ResponseWriter, r *http.Request) {

	wallet, err := app.models.Wallets.Create()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/wallets/%s", wallet.UUID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "wallet successfully created.UUID: " + wallet.UUID.String()}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) changeWalletBalanceHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		WalletId      string  `json:"walletId"`
		OperationType string  `json:"operationType"`
		Amount        float64 `json:"amount"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.WalletId != "", "uuid", "must be provided")
	v.Check(uuid.Validate(input.WalletId) == nil, "uuid", "must be a valid UUID")
	v.Check(input.OperationType != "", "operationType", "must be provided")
	v.Check(input.OperationType == "DEPOSIT" || input.OperationType == "WITHDRAW",
		"operationType", "must be DEPOSIT or WITHDRAW")
	v.Check(input.Amount > 0, "amount", "must be greater than zero")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var wallet *data.Wallet

	wallet, err = app.models.Wallets.Get(uuid.MustParse(input.WalletId))

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if input.OperationType == "WITHDRAW" {
		input.Amount *= -1
	}

	wallet.Balance += input.Amount

	clear(v.Errors)
	v.Check(wallet.Balance >= 0, "Balance", "must be greater than zero")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	attempts := 0
	maxAttempts := 60

	for {
		err = app.models.Wallets.Update(wallet)

		if err == nil {
			break
		} else {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				attempts++
				if attempts >= maxAttempts {
					app.editConflictResponse(w, r)
					return
				}
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"wallet": wallet}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showWalletBalanceHandler(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	stringUUID := params.ByName("uuid")

	v := validator.New()

	v.Check(stringUUID != "", "uuid", "must be provided")
	v.Check(uuid.Validate(stringUUID) == nil, "uuid", "must be a valid UUID")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	wallet, err := app.models.Wallets.Get(uuid.MustParse(stringUUID))

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"wallet": map[string]any{"balance": wallet.Balance}}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteWalletHandler(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	stringUUID := params.ByName("uuid")

	v := validator.New()

	v.Check(stringUUID != "", "uuid", "must be provided")
	v.Check(uuid.Validate(stringUUID) == nil, "uuid", "must be a valid UUID")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err := app.models.Wallets.Delete(uuid.MustParse(stringUUID))

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "wallet successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
