package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/wallet", app.changeWalletBalanceHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/wallets/:uuid", app.showWalletBalanceHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/wallets", app.createWalletHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/wallets/:uuid", app.deleteWalletHandler)

	return router
}
